package pool_test

import (
	"testing"

	"github.com/b-2019-apt-test/divider/pkg/pool"
)

const testPoolJobs = 1000

type TestPoolWorker struct{}

func (w *TestPoolWorker) Process(job interface{}) interface{} {
	return job
}

func TestPool(t *testing.T) {

	p := pool.New(new(TestPoolWorker), 0)
	results := []struct{}{}
	jobs := testPoolJobs

	go func() {
		for i := 0; i < jobs; i++ {
			p.Put(struct{}{})
		}
		p.Close()
	}()

	for result := range p.Consume() {
		results = append(results, result.(struct{}))
	}

	if len(results) < testPoolJobs {
		t.Fatal("not all results received", results)
	}
}
