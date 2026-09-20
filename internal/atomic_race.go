// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build race

package internal

import "sync/atomic"

// The race detector does not instrument this package's assembly. Use sync/atomic,
// whose race implementations in runtime/race_*.s call ThreadSanitizer instead of
// the ordinary internal/runtime/atomic routines. The compiler also disables
// sync/atomic intrinsics under -race so those calls can be intercepted.

// Express the conditional operation using instrumented loads and compare-and-swaps.
func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool) {
	newKey := (new ^ sign) & mask

	for {
		old = atomic.LoadUint64(addr)
		oldKey := (old ^ sign) & mask

		if newKey >= oldKey {
			return old, false
		}

		swapped = atomic.CompareAndSwapUint64(addr, old, new)
		if swapped {
			return old, true
		}
	}
}

func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool) {
	return SwapIfLessUint64(addr, new, mask, sign^mask)
}
