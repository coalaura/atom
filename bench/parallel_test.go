//go:build go1.24

package bench

import (
	"sync/atomic"
	"testing"

	"github.com/coalaura/atom"
)

// Each subbenchmark shares one atomic value across all workers. RunParallel
// requires PB.Next rather than B.Loop and reports aggregate throughput cost.
func BenchmarkParallelLoad(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		value.Store(42)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				value.Load()
			}
		})
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		value.Store(42)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				value.Load()
			}
		})
	})
}

func BenchmarkParallelAdd(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				value.Add(1)
			}
		})
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				value.Add(1)
			}
		})
	})
}

func BenchmarkParallelCASIncrement(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				for {
					previous := value.Load()
					if value.CompareAndSwap(previous, previous+1) {
						break
					}
				}
			}
		})
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				for {
					previous := value.Load()
					if value.CompareAndSwap(previous, previous+1) {
						break
					}
				}
			}
		})
	})
}

func BenchmarkParallelConditionalPair(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[int64]

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				value.SwapIfLess(0)
				value.SwapIfGreater(1)
			}
		})

		b.ReportMetric(2, "calls/op")
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Int64

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(worker *testing.PB) {
			for worker.Next() {
				swapInt64IfLess(&value, 0)
				swapInt64IfGreater(&value, 1)
			}
		})

		b.ReportMetric(2, "calls/op")
	})
}
