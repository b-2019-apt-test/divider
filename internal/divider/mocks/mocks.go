package mocks

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"testing"

	"github.com/b-2019-apt-test/divider/internal/divider"
	"github.com/b-2019-apt-test/divider/pkg/div/godiv"
)

var (
	// FakeLog is used with mocked JobProcessor.
	// By default it uses writer that drops messages.
	FakeLog = log.New(NewFakeWriter(), "", log.LstdFlags)

	// FakeDivider is used with mocked JobProcessor.
	// By default it is godiv.Divider.
	FakeDivider = godiv.Divider

	// TerminalErrorFailFn special error different to divider.NonTerminalError.
	// It is introduced in order to validate behaviour of JobProvider.
	TerminalErrorFailFn = NewErrFailFn(errors.New("terminal error"))

	// NonTerminalErrorFailFn returns empty divider.NonTerminalError
	NonTerminalErrorFailFn = NewErrFailFn(divider.NewNonTerminalError(nil))
)

// NewJobProcessor creates a new mocked JobProcessor.
func NewJobProcessor() *divider.JobProcessor {
	return divider.NewJobProcessor().
		SetJobProvider(NewFakeJobProvider(nil)).
		SetResultReporter(NewFakeResultReporter()).
		SetDivider(FakeDivider).
		SetLogger(FakeLog)
}

// RunTestCases runs provided testing function against mandatory test cases.
func RunTestCases(t *testing.T, run func(*testing.T, TestCase)) {
	run(t, ProcessingError)
	run(t, AllValid)
}

// Validate is a helper for use inside the func provided to RunTestCases.
func Validate(t *testing.T, test TestCase, actual []*divider.JobResult) {
	if !reflect.DeepEqual(test.Results, actual) {
		t.Fatalf("test case %s:\njobs: %s\nexpected: %s\ngot: %s",
			test.Name, pretty(test.Jobs), pretty(test.Results), pretty(actual))
	}
}

func pretty(v interface{}) string {
	s, _ := json.MarshalIndent(v, "", "\t")
	return string(s)
}

// ProviderNonTerminalErrorTest validates proper behaviour of JobProcessor
// on receiving NonTerminalError from JobProvider: it must not stop processing.
//
// Specified JobProvider must return error of type divider.NonTerminalError.
func ProviderNonTerminalErrorTest(t *testing.T, provider divider.JobProvider) {
	if err := NewJobProcessor().SetJobProvider(provider).Start(); err != nil {
		t.Fatal("Non-terminal error caused proccessing fail:", err)
	}
}

// ProviderTerminalErrorTest validates proper termination of JobProcessor on
// receiving TerminalError from JobProvider.
//
// Specified JobProvider must return any error of eny type except
// divider.NonTerminalError error.
func ProviderTerminalErrorTest(t *testing.T, provider divider.JobProvider) {
	if NewJobProcessor().SetJobProvider(provider).Start() == nil {
		t.Fatal("Missed provider terminal error")
	}
}

// TestCase represents test case.
type TestCase struct {
	Name    string
	Jobs    []*divider.Job
	Results []*divider.JobResult
	Err     error
}

// AllValid includes only valid jobs.
var AllValid = TestCase{
	Name: "All valid",
	Jobs: []*divider.Job{
		&divider.Job{Arg1: 2, Arg2: 2, Valid: true},
		&divider.Job{Arg1: 0, Arg2: 2, Valid: true},
		&divider.Job{Arg1: 4, Arg2: -2, Valid: true},
		&divider.Job{Arg1: -4, Arg2: 2, Valid: true},
	},
	Results: []*divider.JobResult{
		&divider.JobResult{ID: 0, Value: 1, Valid: true},
		&divider.JobResult{ID: 1, Value: 0, Valid: true},
		&divider.JobResult{ID: 2, Value: -2, Valid: true},
		&divider.JobResult{ID: 3, Value: -2, Valid: true},
	},
	Err: nil,
}

// ProcessingError includes at least one job that will cause processing error.
var ProcessingError = TestCase{
	Name: "Processing error",
	Jobs: []*divider.Job{
		&divider.Job{Arg1: 2, Arg2: 2, Valid: true},
		&divider.Job{Arg1: 2, Arg2: 0, Valid: true},
		&divider.Job{Arg1: 2, Arg2: 2, Valid: true},
	},
	Results: []*divider.JobResult{
		&divider.JobResult{ID: 0, Value: 1, Valid: true},
		&divider.JobResult{ID: 1, Value: 0, Valid: false},
		&divider.JobResult{ID: 2, Value: 1, Valid: true},
	},
	Err: nil,
}

// InvalidJob includes at least one job that has error and should not be
// processed. The case is only for JobProcessor worker testing.
var InvalidJob = TestCase{
	Name: "Invalid job",
	Jobs: []*divider.Job{
		&divider.Job{Arg1: 2, Arg2: 0, Valid: true},
		&divider.Job{Arg1: 0, Arg2: 0, Valid: false},
		&divider.Job{Arg1: 2, Arg2: 2, Valid: true},
	},
	Results: []*divider.JobResult{
		&divider.JobResult{ID: 0, Value: 0, Valid: false},
		&divider.JobResult{ID: 1, Value: 0, Valid: false},
		&divider.JobResult{ID: 2, Value: 1, Valid: true},
	},
	Err: nil,
}
