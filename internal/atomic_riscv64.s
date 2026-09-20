// Copyright 2014 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

// RISC-V's atomic operations have two bits, aq ("acquire") and rl ("release"),
// which may be toggled on and off. Their precise semantics are defined in
// section 6.3 of the specification, but the basic idea is as follows:
//
//   - If neither aq nor rl is set, the CPU may reorder the atomic arbitrarily.
//     It guarantees only that it will execute atomically.
//
//   - If aq is set, the CPU may move the instruction backward, but not forward.
//
//   - If rl is set, the CPU may move the instruction forward, but not backward.
//
//   - If both are set, the CPU may not reorder the instruction at all.
//
// These four modes correspond to other well-known memory models on other CPUs.
// On ARM, aq corresponds to a dmb ishst, aq+rl corresponds to a dmb ish. On
// Intel, aq corresponds to an lfence, rl to an sfence, and aq+rl to an mfence
// (or a lock prefix).
//
// Go's memory model requires that
//   - if a read happens after a write, the read must observe the write, and
//     that
//   - if a read happens concurrently with a write, the read may observe the
//     write.
// aq is sufficient to guarantee this, so that's what we use here. (This jibes
// with ARM, which uses dmb ishst.)

#include "textflag.h"

// func Cas64(ptr *uint64, old, new uint64) bool
TEXT ·Cas64(SB), NOSPLIT, $0-25
	MOV	ptr+0(FP), A0
	MOV	old+8(FP), A1
	MOV	new+16(FP), A2
cas_again:
	LRD	(A0), A3
	BNE	A3, A1, cas_fail
	SCD	A2, (A0), A4
	BNE	A4, ZERO, cas_again
	MOV	$1, A0
	MOVB	A0, ret+24(FP)
	RET
cas_fail:
	MOVB	ZERO, ret+24(FP)
	RET

// func Load64(ptr *uint64) uint64
TEXT ·Load64(SB),NOSPLIT|NOFRAME,$0-16
	MOV	ptr+0(FP), A0
	LRD	(A0), A0
	MOV	A0, ret+8(FP)
	RET

// func Store64(ptr *uint64, val uint64)
TEXT ·Store64(SB), NOSPLIT, $0-16
	MOV	ptr+0(FP), A0
	MOV	val+8(FP), A1
	AMOSWAPD A1, (A0), ZERO
	RET

// func Xchg64(ptr *uint64, new uint64) uint64
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOV	ptr+0(FP), A0
	MOV	new+8(FP), A1
	AMOSWAPD A1, (A0), A1
	MOV	A1, ret+16(FP)
	RET

// func Xadd64(ptr *uint64, delta int64) uint64
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOV	ptr+0(FP), A0
	MOV	delta+8(FP), A1
	AMOADDD A1, (A0), A2
	ADD	A2, A1, A0
	MOV	A0, ret+16(FP)
	RET

// func Or64(ptr *uint64, val uint64) uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOV	ptr+0(FP), A0
	MOV	val+8(FP), A1
	AMOORD	A1, (A0), A2
	MOV	A2, ret+16(FP)
	RET

// func And64(ptr *uint64, val uint64) uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOV	ptr+0(FP), A0
	MOV	val+8(FP), A1
	AMOANDD	A1, (A0), A2
	MOV	A2, ret+16(FP)
	RET

// func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Atomically:
//	old = *addr
//	swapped = (new^sign)&mask < (old^sign)&mask
//	if swapped {
//		*addr = new
//	}
//	return old, swapped
TEXT ·SwapIfLessUint64(SB), NOSPLIT, $0-41
	MOV	addr+0(FP), A0
	MOV	new+8(FP), A1
	MOV	mask+16(FP), A2
	MOV	sign+24(FP), A3
	XOR	A3, A1, A4
	AND	A2, A4
	// Go's LRD/SCD include acquire/release ordering. Keep the reservation
	// loop short and free of other memory accesses for forward progress.
retry:
	LRD	(A0), A5
	XOR	A3, A5, A6
	AND	A2, A6
	BGEU	A4, A6, unchanged
	SCD	A1, (A0), A7
	BNE	A7, ZERO, retry
	MOV	$1, A7
	MOVB	A7, swapped+40(FP)
	MOV	A5, old+32(FP)
	RET
unchanged:
	MOVB	ZERO, swapped+40(FP)
	MOV	A5, old+32(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-41
	MOV	sign+24(FP), A0
	MOV	mask+16(FP), A1
	XOR	A1, A0
	MOV	A0, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)
