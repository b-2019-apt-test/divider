package jsonprov

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"github.com/b-2019-apt-test/divider/internal/divider"
)

// JSONJobDecoder implements divider.JobProvider. As implied, JSONJobDecoder
// decodes JSON stream of job objects.
type JSONJobDecoder struct {
	dec *json.Decoder
}

// New creates a new JSONJobDecoder with io.Reader.
func New(r io.Reader) (*JSONJobDecoder, error) {
	d := &JSONJobDecoder{dec: json.NewDecoder(r)}
	if _, err := d.dec.Token(); err != nil && err == io.EOF {
		return nil, err
	}
	return d, nil
}

// More reports whether there is another element in the current array or object
// being parsed.
func (d *JSONJobDecoder) More() bool {
	return d.dec.More()
}

// Next reads the next JSON-encoded job from its input.
func (d *JSONJobDecoder) Next(job *divider.Job) (err error) {

	nilJob := &struct{ Arg1, Arg2 *int }{}
	if err = d.dec.Decode(nilJob); err != nil {
		switch typedErr := err.(type) {
		case *json.UnmarshalTypeError:
			return divider.NewNonTerminalError(err)
		case *json.SyntaxError:
			return newJSONSyntaxErrorWithOffset(typedErr)
		}
		return
	}

	if nilJob.Arg1 == nil {
		return divider.ErrArg1Missing
	}
	if nilJob.Arg2 == nil {
		return divider.ErrArg2Missing
	}

	job.Arg1, job.Arg2 = *nilJob.Arg1, *nilJob.Arg2
	job.Valid = true

	return nil
}

// newJSONSyntaxErrorWithOffset is convenience wrapper for json.SyntaxError
// that also provides the syntax error offset.
func newJSONSyntaxErrorWithOffset(e *json.SyntaxError) error {
	return errors.New("JSON syntax error at offset " +
		strconv.FormatInt(e.Offset, 10) + ": " + e.Error())
}
