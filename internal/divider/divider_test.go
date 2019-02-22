package divider_test

import (
	"bytes"
	"errors"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/b-2019-apt-test/divider/internal/divider"
	"github.com/b-2019-apt-test/divider/internal/divider/extra"
	"github.com/b-2019-apt-test/divider/internal/divider/ordinary"
)

type testCase struct {
	jobs    string
	results string
	failed  bool
}

var testCases = []testCase{
	EmptyFile,
	EmptyJSONArray,
	EndlessJSONArray,
	AllValid,
	UnexpectedEOF,
	InvalidSyntax,
	InvalidType,
	FieldUnknown,
	FieldDuplicate,
	Int32Overflow,
	ZeroDivision,
}

var testLog = log.New(NewFakeWriter(), "", log.LstdFlags)

// BrokenWriter returns specified error after n bytes written.
type BrokenWriter struct {
	n, c int
	err  error
}

// FakeWriter does not write anything.
type FakeWriter struct{}

var ErrBrokenWriter = errors.New("Writer is broken")

func NewBrokenWriter(n int, err error) *BrokenWriter {
	return &BrokenWriter{n: n, err: err}
}

func (w *BrokenWriter) Write(b []byte) (int, error) {
	w.c += len(b)
	if w.c > w.n {
		return 0, w.err
	}
	return len(b), nil
}

func NewFakeWriter() *FakeWriter {
	return new(FakeWriter)
}

func (w *FakeWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestJobsWithOrdinaryDivider(t *testing.T) {
	// t.SkipNow()
	testJobs(t, ordinary.NewDivider())
}

func TestJobsWithExtraDivider(t *testing.T) {
	// t.SkipNow()
	testJobs(t, extra.NewDivider())
}

func testJobs(t *testing.T, d divider.Divider) {
	for i, c := range testCases {
		actual, err := runTestJobs(c.jobs, d)
		if err == nil && c.failed {
			t.Fatalf("case %d: missed fail, jobs: %v", i, c.jobs)
		} else if err != nil && !c.failed {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.results, actual) {
			t.Fatalf("case %d:\njobs: %v\nexpected:\n%v\ngot:\n%v",
				i, c.jobs, c.results, actual)
		}
	}
}

func runTestJobs(jobs string, d divider.Divider) (results string, err error) {
	var buf bytes.Buffer
	err = divider.NewJobProcessor().
		SetJobReader(strings.NewReader(jobs)).
		SetResultWriter(&buf).
		SetLogger(testLog).
		SetDivider(d).
		Start()
	if err != nil {
		return
	}
	return buf.String(), nil
}

func TestStartUnconfiguredJobProcessor(t *testing.T) {
	proc := divider.NewJobProcessor().
		SetWorkersCount(2).
		SetPoolSize(4)

	if err := proc.Start(); err != divider.ErrJobReaderNotSpecified {
		t.Fatalf("unconfigured job reader ignored: %v", err)
	}
	proc.SetJobReader(strings.NewReader(AllValid.jobs))

	if err := proc.Start(); err != divider.ErrResultWriterNotSpecified {
		t.Fatalf("unconfigured result writer ignored: %v", err)
	}
	var writer bytes.Buffer
	proc.SetResultWriter(&writer)

	if err := proc.Start(); err != divider.ErrLoggerNotSpecified {
		t.Fatalf("unconfigured logger ignored: %v", err)
	}
	proc.SetLogger(testLog)

	if err := proc.Start(); err != divider.ErrDividerNotSpecified {
		t.Fatalf("unconfigured divider ignored: %v", err)
	}
	proc.SetDivider(extra.NewDivider())

	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}
}

func TestStopJobProcessor(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	var buf bytes.Buffer
	proc := divider.NewJobProcessor().
		SetDivider(extra.NewDivider()).
		SetLogger(testLog).
		SetJobReader(strings.NewReader(AllValid.jobs)).
		SetResultWriter(&buf)

	proc.Stop()

	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}

	proc.Stop()
	proc.Stop()
}

func TestBrokenWriterImediate(t *testing.T) {
	testBrokenWriter(t, 0)
}

func TestBrokenWriterDelayed(t *testing.T) {
	testBrokenWriter(t, len(AllValid.results)-1)
}

func testBrokenWriter(t *testing.T, n int) {
	err := divider.NewJobProcessor().
		SetJobReader(strings.NewReader(AllValid.jobs)).
		SetResultWriter(NewBrokenWriter(n, ErrBrokenWriter)).
		SetDivider(extra.NewDivider()).
		SetLogger(testLog).
		Start()
	if err != ErrBrokenWriter {
		t.Fatalf("expected ErrBrokenWriter, got: %v", err)
	}
}

var (
	EmptyFile = testCase{
		jobs:   ``,
		failed: false,
		results: `id,value,valid
`,
	}

	EmptyJSONArray = testCase{
		jobs:   `[]`,
		failed: false,
		results: `id,value,valid
`,
	}

	EndlessJSONArray = testCase{
		jobs:   `[`,
		failed: false,
		results: `id,value,valid
`,
	}

	AllValid = testCase{
		jobs: `[
	{
		"arg1": 4,
		"arg2": 2
	},
	{
		"arg1": 128,
		"arg2": 16
	},
	{
		"arg1": -128,
		"arg2": 16
	},
	{
		"arg1": 128,
		"arg2": -16
	}
]
`,
		failed: false,
		results: `id,value,valid
0,2,true
1,8,true
2,-8,true
3,-8,true
`,
	}

	InvalidType = testCase{
		jobs: `[
			{"arg1": 4,		"arg2": 2},
			{"arg1": "16",	"arg2": 16}
		]`,
		failed: false,
		results: `id,value,valid
0,2,true
1,0,false
`,
	}

	InvalidSyntax = testCase{
		jobs:   `[{"arg1":,}]`,
		failed: true,
	}

	UnexpectedEOF = testCase{
		jobs:   `[{"arg1"`,
		failed: true,
	}

	FieldUnknown = testCase{
		jobs:   `[{"arg1": 6, "arg2": 2, "arg3": 3}]`,
		failed: false,
		results: `id,value,valid
0,0,false
`,
	}

	FieldDuplicate = testCase{
		jobs:   `[{"arg1": 6, "arg2": 2, "arg2": 3}]`,
		failed: false,
		results: `id,value,valid
0,2,true
`,
	}

	Int32Overflow = testCase{
		jobs:   `[{"arg1": 9223372036854775807, "arg2": 1}]`,
		failed: false,
		results: `id,value,valid
0,0,false
`,
	}

	ZeroDivision = testCase{
		jobs: `[
			{"arg1": 1, "arg2": 0},
			{"arg1": 0, "arg2": 1}	
		]`,
		failed: false,
		results: `id,value,valid
0,0,false
1,0,true
`,
	}
)
