package div

import "errors"

// Divider performs decimal division where a is dividend and b is divisor.
// You still can't divide by zero: in this case Divider returns zero and
// ErrDivZero error.
type Divider interface {
	Div(int, int) (int, error)
}

// ErrDivZero specifies that divisor equals zero.
var ErrDivZero = errors.New("Division by zero")
