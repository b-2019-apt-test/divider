package ordinary

import "errors"

// ErrNilDivisor specifies that divisor equals zero.
var ErrNilDivisor = errors.New("Division by zero")

// Divider performs decimal division with built-in means.
type Divider struct{}

// NewDivider creates new Divider.
func NewDivider() *Divider {
	return new(Divider)
}

// Div divides a by b with rounding down.
func (d *Divider) Div(a, b int32) (int32, error) {
	if b == 0 {
		return 0, ErrNilDivisor
	}
	return a / b, nil
}
