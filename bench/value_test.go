//go:build go1.24

package bench

import (
	"sync/atomic"
	"testing"

	"github.com/coalaura/atom"
)

type payload struct {
	sequence uint64
	stamp    uint64
}

func BenchmarkValuePointerLoad(b *testing.B) {
	var item payload

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[*payload]

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			value.Load()
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			_ = value.Load().(*payload)
		}
	})
}

func BenchmarkValuePointerStore(b *testing.B) {
	var item payload

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[*payload]

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			value.Store(&item)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			value.Store(&item)
		}
	})
}

func BenchmarkValuePointerSwap(b *testing.B) {
	var item payload

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[*payload]

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			value.Swap(&item)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			_ = value.Swap(&item).(*payload)
		}
	})
}

func BenchmarkValuePointerCASChanged(b *testing.B) {
	var (
		first  payload
		second payload
	)

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[*payload]

		previous := &first
		next := &second
		value.Store(previous)
		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(previous, next)
			previous, next = next, previous
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		previous := &first
		next := &second
		value.Store(previous)
		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(previous, next)
			previous, next = next, previous
		}
	})
}

func BenchmarkValuePointerCASRejected(b *testing.B) {
	var (
		first  payload
		second payload
	)

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[*payload]

		value.Store(&first)
		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(&second, &first)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		value.Store(&first)
		b.ReportAllocs()

		for b.Loop() {
			value.CompareAndSwap(&second, &first)
		}
	})
}

func BenchmarkValueStructLoad(b *testing.B) {
	item := payload{sequence: 1000, stamp: 2000}

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[payload]

		value.Store(item)
		b.ReportAllocs()

		for b.Loop() {
			value.Load()
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		value.Store(item)
		b.ReportAllocs()

		for b.Loop() {
			_ = value.Load().(payload)
		}
	})
}

func BenchmarkValueStructStore(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[payload]

		item := payload{sequence: 1000, stamp: 2000}
		value.Store(item)
		b.ReportAllocs()

		for b.Loop() {
			item.sequence++
			value.Store(item)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		item := payload{sequence: 1000, stamp: 2000}
		value.Store(item)
		b.ReportAllocs()

		for b.Loop() {
			item.sequence++
			value.Store(item)
		}
	})
}

func BenchmarkValueStructSwap(b *testing.B) {
	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[payload]

		item := payload{sequence: 1000, stamp: 2000}
		value.Store(item)
		b.ReportAllocs()

		for b.Loop() {
			item.sequence++
			value.Swap(item)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		item := payload{sequence: 1000, stamp: 2000}
		value.Store(item)
		b.ReportAllocs()

		for b.Loop() {
			item.sequence++
			_ = value.Swap(item).(payload)
		}
	})
}

func BenchmarkValueSwapIfFuncChanged(b *testing.B) {
	var (
		first  payload
		second payload
	)

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[*payload]

		previous := &first
		next := &second
		value.Store(previous)
		b.ReportAllocs()

		for b.Loop() {
			value.SwapIfFunc(next, acceptPointer)
			previous, next = next, previous
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		previous := &first
		next := &second
		value.Store(previous)
		b.ReportAllocs()

		for b.Loop() {
			swapPointerIfFunc(&value, next, acceptPointer)
			previous, next = next, previous
		}
	})
}

func BenchmarkValueSwapIfFuncRejected(b *testing.B) {
	var item payload

	b.Run("Atom", func(b *testing.B) {
		var value atom.Value[*payload]

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			value.SwapIfFunc(&item, rejectPointer)
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		var value atomic.Value

		value.Store(&item)
		b.ReportAllocs()

		for b.Loop() {
			swapPointerIfFunc(&value, &item, rejectPointer)
		}
	})
}

func swapPointerIfFunc(value *atomic.Value, next *payload, predicate func(*payload) bool) (*payload, bool) {
	for {
		previous := value.Load().(*payload)
		if !predicate(previous) {
			return previous, false
		}

		if value.CompareAndSwap(previous, next) {
			return previous, true
		}
	}
}

func acceptPointer(previous *payload) bool {
	return previous != nil
}

func rejectPointer(previous *payload) bool {
	return previous == nil
}
