package divider

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"github.com/b-2019-apt-test/divider/pkg/div"
	"github.com/b-2019-apt-test/divider/pkg/pool"
)

// Job to be processed in accordance to the aptitude test requirements.
type Job struct {
	Arg1, Arg2 int
	Valid      bool
}

// JobResult represents an outcome of a particular job processing. ID member
// is a sequential number of the corresponding job. Valid member specifies
// whether the job was processed correctly.
type JobResult struct {
	ID    uint64
	Value int
	Valid bool
}

// JobProcessor does all the job - processes jobs in accordance to the aptitude
// test requirements.
type JobProcessor struct {
	d    div.Divider
	prov JobProvider
	rrep ResultReporter
	log  *log.Logger

	pool *pool.Pool
	wc   uint

	cancel context.CancelFunc

	c uint64
	p uint64
}

// JobProvider supplies new jobs to JobProcessor. JobProvider must return error
// of NonTerminalError type if provided jobs was is invalid.
type JobProvider interface {
	Next(*Job) error
	More() bool
}

// ResultReporter delievers results of job processing.
type ResultReporter interface {
	Report(*JobResult) error
}

// NonTerminalError specifies that JobProcessor must skip processing of the job.
// The error should only raised by JobProvider.
type NonTerminalError struct {
	err error
}

// NewNonTerminalError creates new NonTerminalError that wraps provided one.
func NewNonTerminalError(err error) *NonTerminalError {
	return &NonTerminalError{err}
}

func (e *NonTerminalError) Error() string {
	return e.err.Error()
}

type worker struct {
	div.Divider
	log *log.Logger
}

type workerTask struct {
	job    *Job
	result *JobResult
}

var (
	// ErrJobProviderNotSpecified returns if JobProcessor started without
	// specified JobProvider.
	ErrJobProviderNotSpecified = errors.New("JobProvider not specified")

	// ErrResultReporterNotSpecified returns if JobProcessor started without
	// specified ResultReporter.
	ErrResultReporterNotSpecified = errors.New("ResultReporter not specified")

	// ErrDividerNotSpecified returns if JobProcessor started without
	// specified div.Divider for processing jobs.
	ErrDividerNotSpecified = errors.New("Divider not specified")

	// ErrLoggerNotSpecified returned if JobProcessor started without
	// specified log.Logger.
	ErrLoggerNotSpecified = errors.New("Logger not specified")

	// ErrArg1Missing specifies that a job does not have field "arg1"
	ErrArg1Missing = NewNonTerminalError(errors.New(`"arg1" field missing`))

	// ErrArg2Missing specifies that a job does not have field "arg2"
	ErrArg2Missing = NewNonTerminalError(errors.New(`"arg2" field missing`))
)

// NewJobProcessor creates a new JobProcessor.
func NewJobProcessor() *JobProcessor {
	return new(JobProcessor)
}

func (p *JobProcessor) newWorker() *worker {
	return &worker{
		Divider: p.d,
		log:     p.log}
}

func (p *JobProcessor) newWorkerTask() *workerTask {
	task := &workerTask{
		job:    &Job{},
		result: &JobResult{ID: p.c}}
	p.c++
	return task
}

// Process implements pool.Worker.
func (w *worker) Process(x interface{}) interface{} {

	task := x.(*workerTask)
	if !task.job.Valid {
		return task.result
	}

	var err error
	task.result.Value, err = w.Div(task.job.Arg1, task.job.Arg2)
	if err != nil {
		w.log.Printf("Job %d processed with error: %v", task.result.ID, err)
	} else {
		task.result.Valid = true
	}

	return task.result
}

// Start initializes worker pool and begins processing of jobs.
func (p *JobProcessor) Start() error {

	if p.prov == nil {
		return ErrJobProviderNotSpecified
	}
	if p.rrep == nil {
		return ErrResultReporterNotSpecified
	}
	if p.log == nil {
		return ErrLoggerNotSpecified
	}
	if p.d == nil {
		return ErrDividerNotSpecified
	}

	p.log.Println("Processing started.")
	p.pool = pool.New(p.newWorker(), p.wc)

	done := make(chan bool)
	go p.reportResults(done)

	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	err := p.enqueueJobs(ctx)

	// We should not wait for result reporter to finish if
	// an error has occurred on the job provider side.
	// The 'done' chan can potentially leak in that case
	// but it is assumed that the program itself exits on
	// a JobProcessor error.
	if err != nil {
		return err
	}

	<-done
	close(done)

	return nil
}

// Stop cancels reading, what, in turn, closes the worker pool and stops
// writing results.
func (p *JobProcessor) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

// Processed returns processed jobs counter.
func (p *JobProcessor) Processed() uint64 {
	return atomic.LoadUint64(&p.p)
}

// enqueueJobs puts jobs received from JobProvider to worker pool.
func (p *JobProcessor) enqueueJobs(ctx context.Context) error {

	defer p.pool.Close()
	var task *workerTask

	for p.prov.More() {
		select {
		case <-ctx.Done():
			return nil
		default:
			task = p.newWorkerTask()
			if err := p.prov.Next(task.job); err != nil {
				if _, ok := err.(*NonTerminalError); !ok {
					return err
				}
				p.log.Printf("Job %d processing error: %v", task.result.ID, err)
			}
			p.pool.Put(task)
		}
	}

	return nil
}

func (p *JobProcessor) reportResults(done chan bool) {
	for result := range p.pool.Consume() {
		if err := p.rrep.Report(result.(*JobResult)); err != nil {
			// Do not try to handle error.
			// If we can't write results, we must exit immediately.
			// Perhaps it worth to log a report on work, even in
			// case of failure.
			p.log.Fatalf("Unable to write report: %v", err)
		}
		atomic.AddUint64(&p.p, 1)
	}
	done <- true
}

// SetJobProvider specifies a provider of jobs to be processed.
func (p *JobProcessor) SetJobProvider(prov JobProvider) *JobProcessor {
	p.prov = prov
	return p
}

// SetResultReporter specifies a writer that will be used with JobProcessor to
// write processing results.
func (p *JobProcessor) SetResultReporter(rrep ResultReporter) *JobProcessor {
	p.rrep = rrep
	return p
}

// SetLogger specifies a logger to be used with JobProcessor and its workers.
func (p *JobProcessor) SetLogger(l *log.Logger) *JobProcessor {
	p.log = l
	return p
}

// SetWorkersCount specifies count of workers to be created to handle jobs.
func (p *JobProcessor) SetWorkersCount(workers uint) *JobProcessor {
	p.wc = workers
	return p
}

// SetDivider sets division method.
func (p *JobProcessor) SetDivider(d div.Divider) *JobProcessor {
	p.d = d
	return p
}
