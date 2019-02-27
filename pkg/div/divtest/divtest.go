package divtest

import (
	"testing"

	"github.com/b-2019-apt-test/divider/pkg/div"
)

type testCase struct {
	a, b, x int
	err     error
}

var testCases = []testCase{
	{1, 0, 0, div.ErrDivZero},
	{0, 1, 0, nil},
	{1, 2, 0, nil},
	{1, 1, 1, nil},
	{1, -1, -1, nil},
	{-1, 1, -1, nil},
	{-1, -1, 1, nil},
	{4, 2, 2, nil},
}

// Cases runs divider through test cases.
func Cases(t *testing.T, d div.Divider) {
	for _, test := range testCases {
		x, err := d.Div(test.a, test.b)
		if err != nil && test.err == nil {
			t.Fatalf("case: %#v\nunexpected err: %v", test, err)
		}
		if err == nil && test.err != nil {
			t.Fatalf("case: %#v\nerror miss: %v\ngot: %v", test, test.err, err)
		}
		if x != test.x {
			t.Fatalf("case: %#v\nexpected result: %v\ngot: %v", test, test.x, x)
		}
	}
}

// Benchmark performs simple benchmark of the divider.
func Benchmark(b *testing.B, d div.Divider) {
	b.Run("Once", func(b *testing.B) {
		d.Div(1e9, 1)
	})
	b.Run("b.N", func(b *testing.B) {
		for i := 1; i < b.N; i++ {
			d.Div(1e9, i)
		}
	})
}

// BenchmarkParallel
func BenchmarkParallel(b *testing.B, d div.Divider) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d.Div(1e9, 1)
		}
	})
}
