package csvrep

import (
	"io"
	"strconv"

	"github.com/b-2019-apt-test/divider/internal/divider"
)

// CSVResultReporter produces CSV-formatted output of processing results and
// writes it with provided Writer.
type CSVResultReporter struct {
	w io.Writer
}

// New creates a new CSVResultReporter.
func New(w io.Writer) (*CSVResultReporter, error) {
	r := &CSVResultReporter{w}
	if _, err := r.w.Write([]byte("id,value,valid\n")); err != nil {
		return nil, err
	}
	return r, nil
}

// Report writes a CSV-formatted line for the job result.
func (r *CSVResultReporter) Report(result *divider.JobResult) error {
	_, err := r.w.Write([]byte((strconv.FormatUint(result.ID, 10) + "," +
		strconv.FormatInt(int64(result.Value), 10) + "," +
		strconv.FormatBool(result.Valid) + "\n")))
	return err
}
