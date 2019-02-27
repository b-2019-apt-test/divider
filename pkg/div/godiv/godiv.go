package godiv

import "github.com/b-2019-apt-test/divider/pkg/div"

// Divider performs decimal division with built-in means.
var Divider = divider{}

type divider struct{}

// Div divides a by b with rounding down.
func (d divider) Div(a, b int) (int, error) {
	if b == 0 {
		return 0, div.ErrDivZero
	}
	return a / b, nil
}
