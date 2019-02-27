package calldiv

import (
	"syscall"

	"github.com/b-2019-apt-test/divider/pkg/div"
)

var (
	math    = syscall.NewLazyDLL(mathDLLName)
	divProc = math.NewProc("Div")
)

// Divider performs decimal division by calling external lib (math.dll).
var Divider = divider{}

type divider struct{}

// Div divides a by b with rounding down. The actual returned value is of type
// int32 regardles of the current platform arch.
func (d divider) Div(a, b int) (int, error) {
	if b == 0 {
		return 0, div.ErrDivZero
	}
	r, _, err := divProc.Call(uintptr(a), uintptr(b))
	if uintptr(err.(syscall.Errno)) != 0 {
		return 0, err
	}
	return int(int32(r)), nil
}
