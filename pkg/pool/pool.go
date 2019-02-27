package pool

import "sync"

// Worker is a dummy pool worker which processes some jobs and produces
// some results.
type Worker interface {
	Process(job interface{}) (result interface{})
}

// Pool represents a poor man worker pool.
type Pool struct {
	jobs chan interface{}
	wg   sync.WaitGroup
	c    chan interface{}
}

// New creates a pool with wc count of workers.
func New(w Worker, wc uint) *Pool {

	if wc == 0 {
		wc = 1
	}

	pool := &Pool{
		jobs: make(chan interface{}, wc),
		c:    make(chan interface{})}

	for wc > 0 {
		pool.wg.Add(1)
		wc--
		go pool.run(w)
	}

	return pool
}

// Close makes the pool to stop processing of incoming jobs. All buffered
// jobs will be processed and only then the pool will close consumer chan and
// finish the work. No jobs must be put into the pool after the close.
func (p *Pool) Close() {
	close(p.jobs)
	p.wg.Wait()
	close(p.c)
}

// Put adds the job to the pool. The call will be blocked,
// if all workers are busy.
func (p *Pool) Put(job interface{}) {
	p.jobs <- job
}

// Consume is a convenience wrapper for result channel of the pool.
func (p *Pool) Consume() <-chan interface{} {
	return p.c
}

func (p *Pool) run(w Worker) {
	for job := range p.jobs {
		p.c <- w.Process(job)
	}
	p.wg.Done()
}
