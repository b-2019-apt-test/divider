package godiv_test

import (
	"testing"

	"github.com/b-2019-apt-test/divider/pkg/div/divtest"
	"github.com/b-2019-apt-test/divider/pkg/div/godiv"
)

func TestCasesGoDiv(t *testing.T) {
	divtest.Cases(t, godiv.Divider)
}

func BenchmarkGoDiv(b *testing.B) {
	divtest.Benchmark(b, godiv.Divider)
}

func BenchmarkGoDivParallel(b *testing.B) {
	divtest.BenchmarkParallel(b, godiv.Divider)
}
