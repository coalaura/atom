// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm && !race

package internal

import (
	"github.com/coalaura/atom/internal/cpu"
	"unsafe"
)

const (
	offsetARMHasV7Atomics = unsafe.Offsetof(cpu.ARM.HasV7Atomics)
)

type spinlock struct {
	v uint32
}

//go:nosplit
func (l *spinlock) lock() {
	for {
		if Cas(&l.v, 0, 1) {
			return
		}
	}
}

//go:nosplit
func (l *spinlock) unlock() {
	Store(&l.v, 0)
}

var locktab [57]struct {
	l   spinlock
	pad [cpu.CacheLinePadSize - unsafe.Sizeof(spinlock{})]byte
}

func addrLock(addr *uint64) *spinlock {
	return &locktab[(uintptr(unsafe.Pointer(addr))>>3)%uintptr(len(locktab))].l
}

//go:noescape
func Store(addr *uint32, v uint32)

//go:nosplit
func goCas64(addr *uint64, old, new uint64) bool {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var ok bool
	addrLock(addr).lock()
	if *addr == old {
		*addr = new
		ok = true
	}
	addrLock(addr).unlock()
	return ok
}

//go:nosplit
func goXadd64(addr *uint64, delta int64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr + uint64(delta)
	*addr = r
	addrLock(addr).unlock()
	return r
}

//go:nosplit
func goXchg64(addr *uint64, v uint64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr
	*addr = v
	addrLock(addr).unlock()
	return r
}

//go:nosplit
func goLoad64(addr *uint64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr
	addrLock(addr).unlock()
	return r
}

//go:nosplit
func goStore64(addr *uint64, v uint64) {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	addrLock(addr).lock()
	*addr = v
	addrLock(addr).unlock()
}

//go:nosplit
func armcas(ptr *uint32, old, new uint32) bool

//go:noescape
func Cas64(addr *uint64, old, new uint64) bool

//go:noescape
func Xadd64(addr *uint64, delta int64) uint64

//go:noescape
func Xchg64(addr *uint64, v uint64) uint64

//go:noescape
func Load64(addr *uint64) uint64

//go:noescape
func Store64(addr *uint64, v uint64)

// Local conditional-swap support for the pre-v7 ARM lock fallback.

//go:noescape
func swapIfLessLocked(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)

// lockSwap64 joins the upstream striped lock used on pre-v7 ARM. The assembly
// entry point has already checked alignment; fault on nil before taking a lock.
// The caller releases the lock with Store(&lock.v, 0), as spinlock.unlock does.
//
//go:nosplit
func lockSwap64(addr *uint64) *spinlock {
	_ = *addr

	valueLock := addrLock(addr)
	valueLock.lock()

	return valueLock
}
