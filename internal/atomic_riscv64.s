// Copyright 2014 The Go Authors. All rights reserved.
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
