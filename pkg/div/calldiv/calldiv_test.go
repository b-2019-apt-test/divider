package calldiv_test

import (
	"testing"

	"github.com/b-2019-apt-test/divider/pkg/div/calldiv"
	"github.com/b-2019-apt-test/divider/pkg/div/divtest"
)

func TestCasesCallDiv(t *testing.T) {
	divtest.Cases(t, calldiv.Divider)
}

func BenchmarkCallDiv(b *testing.B) {
	divtest.Benchmark(b, calldiv.Divider)
}

func BenchmarkCallDivParallel(b *testing.B) {
	divtest.BenchmarkParallel(b, calldiv.Divider)
}
