package mocks

import "github.com/b-2019-apt-test/divider/internal/divider"

// FakeResultReporter copies processing results and can provide
// them with Results() call.
type FakeResultReporter struct {
	results []*divider.JobResult
	*Failer
}

// NewFakeResultReporter creates a new FakeResultReporter.
func NewFakeResultReporter() *FakeResultReporter {
	return &FakeResultReporter{Failer: NewFailer()}
}

// Report copies provided result.
func (f *FakeResultReporter) Report(result *divider.JobResult) (err error) {
	if err = f.Fail(); err != nil {
		return
	}
	f.results = append(f.results, &(*result))
	return
}

// Results returns previously saved results.
func (f *FakeResultReporter) Results() []*divider.JobResult {
	return f.results
}
