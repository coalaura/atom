// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (ppc64 || ppc64le) && !race

#include "textflag.h"

// uint64 ·Load64(uint64 volatile* ptr)
TEXT ·Load64(SB),NOSPLIT|NOFRAME,$-8-16
	MOVD	ptr+0(FP), R3
	SYNC
	MOVD	0(R3), R3
	CMP	R3, R3, CR7
	BC	4, 30, 1(PC) // bne- cr7,0x4
	ISYNC
	MOVD	R3, ret+8(FP)
	RET

// func	Cas64(ptr *uint64, old, new uint64) bool
// Atomically:
//	if *ptr == old {
//		*ptr = new
//		return true
//	} else {
//		return false
//	}
TEXT ·Cas64(SB), NOSPLIT, $0-25
	MOVD	ptr+0(FP), R3
	MOVD	old+8(FP), R4
	MOVD	new+16(FP), R5
	LWSYNC
cas64_again:
	LDAR	(R3), R6
	CMP	R6, R4
	BNE	cas64_fail
	STDCCC	R5, (R3)
	BNE	cas64_again
	MOVD	$1, R3
	LWSYNC
	MOVB	R3, ret+24(FP)
	RET
cas64_fail:
	LWSYNC
	MOVB	R0, ret+24(FP)
	RET

// uint64 Xadd64(uint64 volatile *val, int64 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R4
	MOVD	delta+8(FP), R5
	LWSYNC
	LDAR	(R4), R3
	ADD	R5, R3
	STDCCC	R3, (R4)
	BNE	-3(PC)
	LWSYNC
	MOVD	R3, ret+16(FP)
	RET

// uint64 Xchg64(ptr *uint64, new uint64)
// Atomically:
//	old := *ptr;
//	*ptr = new;
//	return old;
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R4
	MOVD	new+8(FP), R5
	LWSYNC
	LDAR	(R4), R3
	STDCCC	R5, (R4)
	BNE	-2(PC)
	ISYNC
	MOVD	R3, ret+16(FP)
	RET

TEXT ·Store64(SB), NOSPLIT, $0-16
	MOVD	ptr+0(FP), R3
	MOVD	val+8(FP), R4
	SYNC
	MOVD	R4, 0(R3)
	RET

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R3
	MOVD	val+8(FP), R4
	LWSYNC
again:
	LDAR	(R3), R6
	OR	R4, R6, R7
	STDCCC	R7, (R3)
	BNE	again
	LWSYNC
	MOVD	R6, ret+16(FP)
	RET

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R3
	MOVD	val+8(FP), R4
	LWSYNC
again:
	LDAR	(R3),R6
	AND	R4, R6, R7
	STDCCC	R7, (R3)
	BNE	again
	LWSYNC
	MOVD	R6, ret+16(FP)
	RET
