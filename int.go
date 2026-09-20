// Copyright 2022 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atom

import "sync/atomic"

// An Int is an atomic integer value. The zero value is zero.
//
// Int must not be copied after first use.
type Int[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr] struct {
	v atomic.Uint64
}

// Load atomically loads and returns the value stored in x.
func (x *Int[T]) Load() T { return T(x.v.Load()) }

// Store atomically stores val into x.
func (x *Int[T]) Store(val T) { x.v.Store(uint64(val)) }

// Swap atomically stores new into x and returns the previous value.
func (x *Int[T]) Swap(new T) (old T) { return T(x.v.Swap(uint64(new))) }

// SwapIfLess atomically stores new into x if new is less than the current value.
// It returns the previous value and whether the swap occurred.
// The comparison uses the signedness and width of T. Equal values are not swapped.
func (x *Int[T]) SwapIfLess(new T) (old T, swapped bool) {
	for {
		oldBits := x.v.Load()
		old = T(oldBits)

		if new >= old {
			return old, false
		}

		// Add may leave overflow bits above T's width; CAS must compare the full stored value.
		swapped = x.v.CompareAndSwap(oldBits, uint64(new))
		if swapped {
			return old, true
		}
	}
}

// SwapIfGreater atomically stores new into x if new is greater than the current value.
// It returns the previous value and whether the swap occurred.
// The comparison uses the signedness and width of T. Equal values are not swapped.
func (x *Int[T]) SwapIfGreater(new T) (old T, swapped bool) {
	for {
		oldBits := x.v.Load()
		old = T(oldBits)

		if new <= old {
			return old, false
		}

		// Compare the original bits, not the potentially narrowed old value.
		swapped = x.v.CompareAndSwap(oldBits, uint64(new))
		if swapped {
			return old, true
		}
	}
}

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Int[T]) CompareAndSwap(old, new T) (swapped bool) {
	return x.v.CompareAndSwap(uint64(old), uint64(new))
}

// Add atomically adds delta to x and returns the new value.
func (x *Int[T]) Add(delta T) (new T) { return T(x.v.Add(uint64(delta))) }

// And atomically performs a bitwise AND operation on x using the bitmask
// provided as mask and returns the old value.
func (x *Int[T]) And(mask T) (old T) { return T(x.v.And(uint64(mask))) }

// Or atomically performs a bitwise OR operation on x using the bitmask
// provided as mask and returns the old value.
func (x *Int[T]) Or(mask T) (old T) { return T(x.v.Or(uint64(mask))) }
