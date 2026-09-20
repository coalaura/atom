// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (mips || mipsle) && !race

package internal

import (
	"github.com/coalaura/atom/internal/cpu"
	"unsafe"
)

// TODO implement lock striping
var lock struct {
	state uint32
	pad   [cpu.CacheLinePadSize - 4]byte
}

//go:noescape
func spinLock(state *uint32)

//go:noescape
func spinUnlock(state *uint32)

//go:nosplit
func lockAndCheck(addr *uint64) {
	// ensure 8-byte alignment
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		panicUnaligned()
	}
	// force dereference before taking lock
	_ = *addr

	spinLock(&lock.state)
}

//go:nosplit
func unlock() {
	spinUnlock(&lock.state)
}

//go:nosplit
func Xadd64(addr *uint64, delta int64) (new uint64) {
	lockAndCheck(addr)

	new = *addr + uint64(delta)
	*addr = new

	unlock()
	return
}

//go:nosplit
func Xchg64(addr *uint64, new uint64) (old uint64) {
	lockAndCheck(addr)

	old = *addr
	*addr = new

	unlock()
	return
}

//go:nosplit
func Cas64(addr *uint64, old, new uint64) (swapped bool) {
	lockAndCheck(addr)

	if (*addr) == old {
		*addr = new
		unlock()
		return true
	}

	unlock()
	return false
}

//go:nosplit
func Load64(addr *uint64) (val uint64) {
	lockAndCheck(addr)

	val = *addr

	unlock()
	return
}

//go:nosplit
func Store64(addr *uint64, val uint64) {
	lockAndCheck(addr)

	*addr = val

	unlock()
	return
}

//go:nosplit
func Or64(addr *uint64, val uint64) (old uint64) {
	for {
		old = *addr
		if Cas64(addr, old, old|val) {
			return old
		}
	}
}

//go:nosplit
func And64(addr *uint64, val uint64) (old uint64) {
	for {
		old = *addr
		if Cas64(addr, old, old&val) {
			return old
		}
	}
}
