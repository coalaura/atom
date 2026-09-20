//go:build go1.24

package bench

import (
	"sync/atomic"
	"testing"

	"github.com/coalaura/atom"
)

func BenchmarkUint64Load(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		value.Store(42)
		b.ReportAllocs()

		for b.Loop() {
			value.Load()
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		value.Store(42)
		b.ReportAllocs()

		for b.Loop() {
			value.Load()
		}
	})
}

func BenchmarkUint64Store(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var (
			value atom.Int[uint64]
			next  uint64
		)

		b.ReportAllocs()

		for b.Loop() {
			next++
			value.Store(next)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var (
			value atomic.Uint64
			next  uint64
		)

		b.ReportAllocs()

		for b.Loop() {
			next++
			value.Store(next)
		}
	})
}

func BenchmarkUint64Swap(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var (
			value atom.Int[uint64]
			next  uint64
		)

		b.ReportAllocs()

		for b.Loop() {
			next++
			value.Swap(next)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var (
			value atomic.Uint64
			next  uint64
		)

		b.ReportAllocs()

		for b.Loop() {
			next++
			value.Swap(next)
		}
	})
}

func BenchmarkUint64CASChanged(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var (
			value    atom.Int[uint64]
			previous uint64
		)

		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(previous, previous+1)
			previous++
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var (
			value    atomic.Uint64
			previous uint64
		)

		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(previous, previous+1)
			previous++
		}
	})
}

func BenchmarkUint64CASRejected(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(1, 2)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(1, 2)
		}
	})
}

func BenchmarkUint64Add(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		b.ReportAllocs()

		for b.Loop() {
			value.Add(1)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		b.ReportAllocs()

		for b.Loop() {
			value.Add(1)
		}
	})
}

func BenchmarkUint64And(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		value.Store(^uint64(0))
		b.ReportAllocs()

		for b.Loop() {
			value.And(0x5555555555555555)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		value.Store(^uint64(0))
		b.ReportAllocs()

		for b.Loop() {
			value.And(0x5555555555555555)
		}
	})
}

func BenchmarkUint64Or(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		b.ReportAllocs()

		for b.Loop() {
			value.Or(0x5555555555555555)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		b.ReportAllocs()

		for b.Loop() {
			value.Or(0x5555555555555555)
		}
	})
}

func BenchmarkInt32Load(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[int32]

		value.Store(-42)
		b.ReportAllocs()

		for b.Loop() {
			value.Load()
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Int32

		value.Store(-42)
		b.ReportAllocs()

		for b.Loop() {
			value.Load()
		}
	})
}

func BenchmarkInt32Add(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[int32]

		b.ReportAllocs()

		for b.Loop() {
			value.Add(1)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Int32

		b.ReportAllocs()

		for b.Loop() {
			value.Add(1)
		}
	})
}
