package jsonprov_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/b-2019-apt-test/divider/internal/divider/jsonprov"
	"github.com/b-2019-apt-test/divider/internal/divider/mocks"
)

var nonTerminalCases = []string{
	`[]`,
	`[{ "arg1": "" }]`,
	`[{ "arg1": 1 }]`,
	`[{ "arg2": 1 }]`,
}

var terminalCases = []string{
	`[{ "arg1": }]`,
	`[{ "a`,
}

func newProvider(s string) *jsonprov.JSONJobDecoder {
	provider, _ := jsonprov.New(strings.NewReader(s))
	return provider
}

func TestNewErr(t *testing.T) {
	if _, err := jsonprov.New(strings.NewReader(``)); err == nil {
		t.Fatal("JSON provider should fail if Reader returned io.EOF")
	}
}

func TestTerminalCases(t *testing.T) {
	for _, badInput := range terminalCases {
		mocks.ProviderTerminalErrorTest(t, newProvider(badInput))
	}
}

func TestNonTerminalCases(t *testing.T) {
	for _, invalidInput := range nonTerminalCases {
		mocks.ProviderNonTerminalErrorTest(t, newProvider(invalidInput))
	}
}

func TestCases(t *testing.T) {
	mocks.RunTestCases(t, func(t *testing.T, test mocks.TestCase) {

		jobs, _ := json.Marshal(test.Jobs)
		provider, _ := jsonprov.New(strings.NewReader(string(jobs)))

		reporter := mocks.NewFakeResultReporter()
		proc := mocks.NewJobProcessor().
			SetResultReporter(reporter).
			SetJobProvider(provider)

		if err := proc.Start(); err != test.Err {
			t.Fatalf("expected err %v, got: %v", test.Err, err)
		}

		mocks.Validate(t, test, reporter.Results())
	})
}
