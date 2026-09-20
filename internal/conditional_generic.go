// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (arm || mips || mipsle) && !race

package internal

import "sync/atomic"

// goSwapIfLess64 interoperates with sync/atomic's lock-based 64-bit operations.
// A private lock cannot protect accesses against the standard library's locks.
// Compare keys at the caller's width, but CAS the full observed backing bits.
func goSwapIfLess64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool) {
	newKey := (new ^ sign) & mask

	for {
		old = atomic.LoadUint64(addr)
		if newKey >= (old^sign)&mask {
			return old, false
		}

		swapped = atomic.CompareAndSwapUint64(addr, old, new)
		if swapped {
			return old, true
		}
	}
}
