// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

package atom

import (
	"unsafe"

	"github.com/coalaura/atom/internal"
)

// SwapIfLess atomically stores new into x if new is less than the current value.
// It returns the previous value and whether the swap occurred.
// The comparison uses the signedness and width of T. Equal values are not swapped.
func (x *Int[T]) SwapIfLess(new T) (old T, swapped bool) {
	old = T(internal.SwapIfLessInteger(&x.v, uint64(new), uint8(unsafe.Sizeof(new)), ^T(0) < 0))
	return old, new < old
}

// SwapIfGreater atomically stores new into x if new is greater than the current value.
// It returns the previous value and whether the swap occurred.
// The comparison uses the signedness and width of T. Equal values are not swapped.
func (x *Int[T]) SwapIfGreater(new T) (old T, swapped bool) {
	old = T(internal.SwapIfGreaterInteger(&x.v, uint64(new), uint8(unsafe.Sizeof(new)), ^T(0) < 0))
	return old, new > old
}
