// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 || race

package atom

import (
	"unsafe"

	"github.com/coalaura/atom/internal"
)

// SwapIfLess atomically stores new into x if new is less than the current value.
// It returns the previous value and whether the swap occurred.
// The comparison uses the signedness and width of T. Equal values are not swapped.
func (x *Int[T]) SwapIfLess(new T) (old T, swapped bool) {
	mask := ^uint64(0) >> (64 - unsafe.Sizeof(new)*8)
	sign := uint64(0)

	if ^T(0) < 0 {
		sign = mask ^ (mask >> 1)
	}

	oldBits, swapped := internal.SwapIfLessUint64(&x.v, uint64(new), mask, sign)
	return T(oldBits), swapped
}

// SwapIfGreater atomically stores new into x if new is greater than the current value.
// It returns the previous value and whether the swap occurred.
// The comparison uses the signedness and width of T. Equal values are not swapped.
func (x *Int[T]) SwapIfGreater(new T) (old T, swapped bool) {
	mask := ^uint64(0) >> (64 - unsafe.Sizeof(new)*8)
	sign := uint64(0)

	if ^T(0) < 0 {
		sign = mask ^ (mask >> 1)
	}

	oldBits, swapped := internal.SwapIfGreaterUint64(&x.v, uint64(new), mask, sign)
	return T(oldBits), swapped
}
