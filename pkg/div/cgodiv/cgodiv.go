package cgodiv

/*
#cgo pkg-config: divider
#include "magic.h"
*/
import "C"
import "github.com/b-2019-apt-test/divider/pkg/div"

// Divider performs decimal division using magic library with cgo means.
var Divider = divider{}

type divider struct{}

// Div divides a by b with rounding down.
func (d divider) Div(a, b int) (int, error) {
	if b == 0 {
		return 0, div.ErrDivZero
	}
	return int(C.Div(_Ctype_int(a), _Ctype_int(b))), nil
}
