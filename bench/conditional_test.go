//go:build go1.24

package bench

import (
	"sync/atomic"
	"testing"

	"github.com/coalaura/atom"
)

func BenchmarkConditionalInt64LessChanged(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var (
			value atom.Int[int64]
			next  int64
		)

		b.ReportAllocs()

		for b.Loop() {
			next--
			value.SwapIfLess(next)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var (
			value atomic.Int64
			next  int64
		)

		b.ReportAllocs()

		for b.Loop() {
			next--
			swapInt64IfLess(&value, next)
		}
	})
}

func BenchmarkConditionalInt64GreaterChanged(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var (
			value atom.Int[int64]
			next  int64
		)

		b.ReportAllocs()

		for b.Loop() {
			next++
			value.SwapIfGreater(next)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var (
			value atomic.Int64
			next  int64
		)

		b.ReportAllocs()

		for b.Loop() {
			next++
			swapInt64IfGreater(&value, next)
		}
	})
}

func BenchmarkConditionalUint64LessChanged(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[uint64]

		next := ^uint64(0)
		value.Store(next)
		b.ReportAllocs()

		for b.Loop() {
			next--
			value.SwapIfLess(next)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Uint64

		next := ^uint64(0)
		value.Store(next)
		b.ReportAllocs()

		for b.Loop() {
			next--
			swapUint64IfLess(&value, next)
		}
	})
}

func BenchmarkConditionalUint64GreaterChanged(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var (
			value atom.Int[uint64]
			next  uint64
		)

		b.ReportAllocs()

		for b.Loop() {
			next++
			value.SwapIfGreater(next)
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
			swapUint64IfGreater(&value, next)
		}
	})
}

func BenchmarkConditionalLessRejected(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[int64]

		value.Store(-34)
		b.ReportAllocs()

		for b.Loop() {
			value.SwapIfLess(12)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Int64

		value.Store(-34)
		b.ReportAllocs()

		for b.Loop() {
			swapInt64IfLess(&value, 12)
		}
	})
}

func BenchmarkConditionalGreaterRejected(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[int64]

		value.Store(12)
		b.ReportAllocs()

		for b.Loop() {
			value.SwapIfGreater(-34)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Int64

		value.Store(12)
		b.ReportAllocs()

		for b.Loop() {
			swapInt64IfGreater(&value, -34)
		}
	})
}

func BenchmarkConditionalLessEqual(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[int64]

		value.Store(-34)
		b.ReportAllocs()

		for b.Loop() {
			value.SwapIfLess(-34)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Int64

		value.Store(-34)
		b.ReportAllocs()

		for b.Loop() {
			swapInt64IfLess(&value, -34)
		}
	})
}

func BenchmarkConditionalGreaterEqual(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Int[int64]

		value.Store(-34)
		b.ReportAllocs()

		for b.Loop() {
			value.SwapIfGreater(-34)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Int64

		value.Store(-34)
		b.ReportAllocs()

		for b.Loop() {
			swapInt64IfGreater(&value, -34)
		}
	})
}

// These helpers mirror the public (old, swapped) contract using typed stdlib
// atomics. The load is retried after a failed CAS, including under contention.
func swapInt64IfLess(value *atomic.Int64, next int64) (int64, bool) {
	for {
		previous := value.Load()
		if next >= previous {
			return previous, false
		}

		if value.CompareAndSwap(previous, next) {
			return previous, true
		}
	}
}

func swapInt64IfGreater(value *atomic.Int64, next int64) (int64, bool) {
	for {
		previous := value.Load()
		if next <= previous {
			return previous, false
		}

		if value.CompareAndSwap(previous, next) {
			return previous, true
		}
	}
}

func swapUint64IfLess(value *atomic.Uint64, next uint64) (uint64, bool) {
	for {
		previous := value.Load()
		if next >= previous {
			return previous, false
		}

		if value.CompareAndSwap(previous, next) {
			return previous, true
		}
	}
}

func swapUint64IfGreater(value *atomic.Uint64, next uint64) (uint64, bool) {
	for {
		previous := value.Load()
		if next <= previous {
			return previous, false
		}

		if value.CompareAndSwap(previous, next) {
			return previous, true
		}
	}
}
