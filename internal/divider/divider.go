package divider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/b-2019-apt-test/divider/pkg/pool"
)

// Job to be processed in accordance to the aptitude test requirements.
type Job struct {
	Arg1, Arg2 int32
}

// JobResult represents an outcome of a particular job processing. ID member
// is a sequential number of the corresponding job. Valid member specifies
// whether the job was processed correctly.
type JobResult struct {
	ID    int64
	Value int32
	Valid bool
}

func (j *JobResult) Header() string {
	return fmt.Sprintf("id,value,valid\n")
}

func (j *JobResult) String() string {
	return fmt.Sprintf("%d,%d,%t\n", j.ID, j.Value, j.Valid)
}

// Divider performs decimal division where a is dividend and b is divisor.
type Divider interface {
	Div(a int32, b int32) (int32, error)
}

type JobProcessor struct {
	log    *log.Logger
	queue  *pool.Pool
	wc, bs uint
	d      Divider
	in     io.Reader
	out    io.Writer
	c      int64
	cancel context.CancelFunc
	err    error
}

type worker struct {
	Divider
	log *log.Logger
}

type workerTask struct {
	valid  bool
	job    *Job
	result *JobResult
}

const (
	poolWorkers = 1
	poolBufSize = 1
)

var (
	ErrJobReaderNotSpecified    = errors.New("JobReader not specified")
	ErrResultWriterNotSpecified = errors.New("ResultWriter not specified")
	ErrDividerNotSpecified      = errors.New("Divider not specified")
	ErrLoggerNotSpecified       = errors.New("Logger not specified")
)

func newJSONSyntaxErrorWithOffset(e *json.SyntaxError) error {
	return fmt.Errorf("JSON syntax error at offset %d: %v", e.Offset, e.Error())
}

func NewJobProcessor() *JobProcessor {
	return &JobProcessor{
		wc: poolWorkers,
		bs: poolBufSize}
}

func (p *JobProcessor) newWorker() *worker {
	return &worker{
		Divider: p.d,
		log:     p.log}
}

func (p *JobProcessor) newWorkerTask() (task *workerTask) {
	task = &workerTask{
		job:    &Job{},
		result: &JobResult{ID: p.c},
	}
	p.c++
	return
}

// Process implements pool.Worker.
func (w *worker) Process(x interface{}) interface{} {
	var err error
	task := x.(*workerTask)
	if !task.valid {
		return task.result
	}
	task.result.Value, err = w.Div(task.job.Arg1, task.job.Arg2)
	if err != nil {
		w.log.Printf("job %d (%#v) result invalid: %+v",
			task.result.ID, task.job, err)
	} else {
		task.result.Valid = true
	}
	return task.result
}

func (p *JobProcessor) SetJobReader(r io.Reader) *JobProcessor {
	p.in = r
	return p
}

func (p *JobProcessor) SetResultWriter(w io.Writer) *JobProcessor {
	p.out = w
	return p
}

func (p *JobProcessor) SetLogger(l *log.Logger) *JobProcessor {
	p.log = l
	return p
}

func (p *JobProcessor) SetWorkersCount(workers uint) *JobProcessor {
	p.wc = workers
	return p
}

func (p *JobProcessor) SetPoolSize(size uint) *JobProcessor {
	p.bs = size
	return p
}

func (p *JobProcessor) SetDivider(d Divider) *JobProcessor {
	p.d = d
	return p
}

func (p *JobProcessor) Start() error {

	if p.in == nil {
		return ErrJobReaderNotSpecified
	}
	if p.out == nil {
		return ErrResultWriterNotSpecified
	}
	if p.log == nil {
		return ErrLoggerNotSpecified
	}
	if p.d == nil {
		return ErrDividerNotSpecified
	}

	p.queue = pool.New(p.newWorker(), p.wc, p.bs)

	if _, err := p.out.Write([]byte(new(JobResult).Header())); err != nil {
		return err
	}

	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	done := make(chan interface{})
	go p.writeResults(done)
	err := p.enqueueJobs(ctx)
	<-done

	if err != nil && err != io.EOF {
		return err
	}
	if p.err != nil {
		return p.err
	}

	return nil
}

// Stop cancels reading, what, in turn, closes the worker pool and stops
// writing results.
func (p *JobProcessor) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

// enqueueJobs parses JSON stream from Reader and puts decoded tasks to the
// worker pool.
func (p *JobProcessor) enqueueJobs(ctx context.Context) (err error) {
	defer p.queue.Close()

	dec := json.NewDecoder(p.in)
	dec.DisallowUnknownFields()
	if _, err = dec.Token(); err != nil {
		return err
	}

	var task *workerTask
	for dec.More() {
		select {
		case <-ctx.Done():
			return nil
		default:
			task = p.newWorkerTask()
			if err = dec.Decode(task.job); err != nil {
				if serr, ok := err.(*json.SyntaxError); ok {
					return newJSONSyntaxErrorWithOffset(serr)
				}
				if err == io.ErrUnexpectedEOF {
					return err
				}
				p.log.Printf("job %d decoding error: %v", task.result.ID, err)
			} else {
				task.valid = true
			}
			p.queue.Put(task)
		}
	}

	if _, err = dec.Token(); err != nil {
		return err
	}

	return nil
}

func (p *JobProcessor) writeResults(done chan interface{}) {
	for result := range p.queue.Consume() {
		_, err := p.out.Write([]byte(fmt.Sprint(result.(*JobResult))))
		if err != nil {
			p.err = err
			p.Stop()
			break
		}
	}
	done <- struct{}{}
}
