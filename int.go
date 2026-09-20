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
	_ noCopy
	_ align64
	v uint64
}

// Load atomically loads and returns the value stored in x.
func (x *Int[T]) Load() T { return T(atomic.LoadUint64(&x.v)) }

// Store atomically stores val into x.
func (x *Int[T]) Store(val T) { atomic.StoreUint64(&x.v, uint64(val)) }

// Swap atomically stores new into x and returns the previous value.
func (x *Int[T]) Swap(new T) (old T) { return T(atomic.SwapUint64(&x.v, uint64(new))) }

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Int[T]) CompareAndSwap(old, new T) (swapped bool) {
	return atomic.CompareAndSwapUint64(&x.v, uint64(old), uint64(new))
}

// Add atomically adds delta to x and returns the new value.
func (x *Int[T]) Add(delta T) (new T) { return T(atomic.AddUint64(&x.v, uint64(delta))) }

// And atomically performs a bitwise AND operation on x using the bitmask
// provided as mask and returns the old value.
func (x *Int[T]) And(mask T) (old T) { return T(atomic.AndUint64(&x.v, uint64(mask))) }

// Or atomically performs a bitwise OR operation on x using the bitmask
// provided as mask and returns the old value.
func (x *Int[T]) Or(mask T) (old T) { return T(atomic.OrUint64(&x.v, uint64(mask))) }

// noCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// align64 may be added to structs that must be 64-bit aligned.
type align64 [0]atomic.Uint64
