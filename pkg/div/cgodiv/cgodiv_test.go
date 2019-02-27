package cgodiv_test

import (
	"testing"

	"github.com/b-2019-apt-test/divider/pkg/div/cgodiv"
	"github.com/b-2019-apt-test/divider/pkg/div/divtest"
)

func TestCasesCGODiv(t *testing.T) {
	divtest.Cases(t, cgodiv.Divider)
}

func BenchmarkCGODiv(b *testing.B) {
	divtest.Benchmark(b, cgodiv.Divider)
}

func BenchmarkCGODivParallel(b *testing.B) {
	divtest.BenchmarkParallel(b, cgodiv.Divider)
}
