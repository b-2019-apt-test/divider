package csvrep_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/b-2019-apt-test/divider/internal/divider"
	"github.com/b-2019-apt-test/divider/internal/divider/csvrep"
	"github.com/b-2019-apt-test/divider/internal/divider/mocks"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	if _, err := csvrep.New(&buf); err != nil {
		t.Fatal(err)
	}
}

func TestNewErr(t *testing.T) {
	if _, err := csvrep.New(mocks.NewBrokenWriter(0)); err == nil {
		t.Fatal("CSV result reporter should fail on writer error")
	}
}

func TestCases(t *testing.T) {
	mocks.RunTestCases(t, func(t *testing.T, test mocks.TestCase) {

		var buf bytes.Buffer
		reporter, _ := csvrep.New(&buf)
		proc := mocks.NewJobProcessor().
			SetJobProvider(mocks.NewFakeJobProvider(test.Jobs)).
			SetResultReporter(reporter)

		if err := proc.Start(); err != test.Err {
			t.Fatalf("expected err %v, got: %v", test.Err, err)
		}

		results, err := results(&buf)
		if err != nil {
			t.Fatalf("failed to parse CSV results (%s): %v", buf.String(), err)
		}

		mocks.Validate(t, test, results)
	})
}

func results(r io.Reader) ([]*divider.JobResult, error) {
	csvr := csv.NewReader(r)
	csvr.Comma = ','

	// header
	fields, err := csvr.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %v", err)
	}
	if len(fields) != 3 {
		return nil, fmt.Errorf("not all columns: %v", fields)
	}
	header := []string{"id", "value", "valid"}
	for i := range header {
		if fields[i] != header[i] {
			return nil, fmt.Errorf("column \"%d\" must have name \"%s\"", i, header[i])
		}
	}

	// body
	var results []*divider.JobResult
	for {
		fields, err := csvr.Read()
		if err == io.EOF {
			return results, nil
		}
		if err != nil {
			return nil, err
		}
		if len(fields) != 3 {
			return nil, fmt.Errorf("not all fields: %v", fields)
		}

		id, _ := strconv.Atoi(fields[0])
		value, _ := strconv.Atoi(fields[1])
		valid, _ := strconv.ParseBool(fields[2])
		results = append(results, &divider.JobResult{
			ID:    uint64(id),
			Value: int(value),
			Valid: valid,
		})
	}
}
