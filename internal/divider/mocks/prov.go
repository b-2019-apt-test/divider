package mocks

import "github.com/b-2019-apt-test/divider/internal/divider"

// FakeJobProvider implements divider.JobProvider.
type FakeJobProvider struct {
	jobs []*divider.Job
	*Failer
	n int
}

// NewFakeJobProvider creates a new FakeJobProvider with io.Reader.
func NewFakeJobProvider(jobs []*divider.Job) *FakeJobProvider {
	return &FakeJobProvider{jobs: jobs, Failer: NewFailer()}
}

// More reports whether there is another element in the current array or object
// being parsed.
func (f *FakeJobProvider) More() bool {
	return f.n < len(f.jobs)
}

// Next reads the next job from its list.
func (f *FakeJobProvider) Next(job *divider.Job) error {
	if f.n > len(f.jobs) {
		return nil
	}
	*job = *f.jobs[f.n]
	f.n++
	return f.Fail()
}
