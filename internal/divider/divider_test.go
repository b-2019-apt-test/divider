package divider_test

import (
	"testing"
	"time"

	"github.com/b-2019-apt-test/divider/internal/divider"
	"github.com/b-2019-apt-test/divider/internal/divider/mocks"
)

var runner = func(t *testing.T, test mocks.TestCase) {

	reporter := mocks.NewFakeResultReporter()
	proc := mocks.NewJobProcessor().
		SetJobProvider(mocks.NewFakeJobProvider(test.Jobs)).
		SetResultReporter(reporter)

	if err := proc.Start(); err != test.Err {
		t.Fatalf("expected err %v, got: %v", test.Err, err)
	}

	mocks.Validate(t, test, reporter.Results())
}

func TestCases(t *testing.T) {
	mocks.RunTestCases(t, runner)
}

func TestWorkerSkipsInvalidJob(t *testing.T) {
	runner(t, mocks.InvalidJob)
}

func TestStartUnconfiguredJobProcessor(t *testing.T) {
	proc := divider.NewJobProcessor().SetWorkersCount(10)

	if err := proc.Start(); err != divider.ErrJobProviderNotSpecified {
		t.Fatalf("unconfigured job provider ignored: %v", err)
	}
	proc.SetJobProvider(mocks.NewFakeJobProvider(mocks.AllValid.Jobs))

	if err := proc.Start(); err != divider.ErrResultReporterNotSpecified {
		t.Fatalf("unconfigured result reporter ignored: %v", err)
	}
	proc.SetResultReporter(mocks.NewFakeResultReporter())

	if err := proc.Start(); err != divider.ErrLoggerNotSpecified {
		t.Fatalf("unconfigured logger ignored: %v", err)
	}
	proc.SetLogger(mocks.FakeLog)

	if err := proc.Start(); err != divider.ErrDividerNotSpecified {
		t.Fatalf("unconfigured divider ignored: %v", err)
	}
	proc.SetDivider(mocks.FakeDivider)

	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}
}

func TestJobProcessorProccessedCount(t *testing.T) {

	proc := mocks.NewJobProcessor().
		SetJobProvider(mocks.NewFakeJobProvider(mocks.AllValid.Jobs))

	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}

	if int(proc.Processed()) != len(mocks.AllValid.Jobs) {
		t.Fatal("Invalid number of processed jobs")
	}
}

func TestJobProcessorStop(t *testing.T) {

	var stop time.Time

	// slowed down provider
	provider := mocks.NewFakeJobProvider(mocks.AllValid.Jobs)
	provider.FailEvery(1).FailFn(mocks.StuckFailFn)
	proc := mocks.NewJobProcessor().SetJobProvider(provider)

	go func() {
		time.Sleep(200 * time.Millisecond)
		proc.Stop()
		stop = time.Now()
	}()

	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}

	if (time.Since(stop).Nanoseconds() / 1e9) > 3 {
		t.Fatal("Processor did not stop in expected time.")
	}
}

func TestProviderNonTerminalError(t *testing.T) {
	provider := mocks.NewFakeJobProvider(mocks.AllValid.Jobs)
	provider.FailOn(1).FailFn(mocks.NonTerminalErrorFailFn)
	mocks.ProviderNonTerminalErrorTest(t, provider)
}

func TestProviderTerminalError(t *testing.T) {
	provider := mocks.NewFakeJobProvider(mocks.AllValid.Jobs)
	provider.FailOn(1).FailFn(mocks.TerminalErrorFailFn)
	mocks.ProviderTerminalErrorTest(t, provider)
}
