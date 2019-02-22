//+build windows

package extra

import "syscall"

var (
	math = syscall.NewLazyDLL(mathDLLName)
	div  = math.NewProc("Div")
)

// Divider performs decimal division by calling external lib (math.dll).
type Divider struct{}

// NewDivider creates new Divider.
func NewDivider() *Divider {
	return new(Divider)
}

// Div divides a by b with rounding down. The actual returned value is of type
// int32 regardles of the current platform arch.
func (d *Divider) Div(a, b int32) (int32, error) {
	ret, _, err := div.Call(uintptr(a), uintptr(b))
	if uintptr(err.(syscall.Errno)) != 0 {
		return 0, err
	}
	return int32(ret), nil
}
