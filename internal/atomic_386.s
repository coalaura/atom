// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "textflag.h"
#include "funcdata.h"

// func Cas64(ptr *uint64, old, new uint64) bool
// Atomically:
//	if *ptr == old {
//		*ptr = new
//		return true
//	} else {
//		return false
//	}
TEXT ·Cas64(SB), NOSPLIT, $0-21
	NO_LOCAL_POINTERS
	MOVL	ptr+0(FP), BP
	TESTL	$7, BP
	JZ	2(PC)
	CALL	·panicUnaligned(SB)
	MOVL	old_lo+4(FP), AX
	MOVL	old_hi+8(FP), DX
	MOVL	new_lo+12(FP), BX
	MOVL	new_hi+16(FP), CX
	LOCK
	CMPXCHG8B	0(BP)
	SETEQ	ret+20(FP)
	RET

TEXT ·Xadd64(SB), NOSPLIT, $0-20
	NO_LOCAL_POINTERS
	// no XADDQ so use CMPXCHG8B loop
	MOVL	ptr+0(FP), BP
	TESTL	$7, BP
	JZ	2(PC)
	CALL	·panicUnaligned(SB)
	// DI:SI = delta
	MOVL	delta_lo+4(FP), SI
	MOVL	delta_hi+8(FP), DI
	// DX:AX = *addr
	MOVL	0(BP), AX
	MOVL	4(BP), DX
addloop:
	// CX:BX = DX:AX (*addr) + DI:SI (delta)
	MOVL	AX, BX
	MOVL	DX, CX
	ADDL	SI, BX
	ADCL	DI, CX

	// if *addr == DX:AX {
	//	*addr = CX:BX
	// } else {
	//	DX:AX = *addr
	// }
	// all in one instruction
	LOCK
	CMPXCHG8B	0(BP)

	JNZ	addloop

	// success
	// return CX:BX
	MOVL	BX, ret_lo+12(FP)
	MOVL	CX, ret_hi+16(FP)
	RET

TEXT ·Xchg64(SB),NOSPLIT,$0-20
	NO_LOCAL_POINTERS
	// no XCHGQ so use CMPXCHG8B loop
	MOVL	ptr+0(FP), BP
	TESTL	$7, BP
	JZ	2(PC)
	CALL	·panicUnaligned(SB)
	// CX:BX = new
	MOVL	new_lo+4(FP), BX
	MOVL	new_hi+8(FP), CX
	// DX:AX = *addr
	MOVL	0(BP), AX
	MOVL	4(BP), DX
swaploop:
	// if *addr == DX:AX
	//	*addr = CX:BX
	// else
	//	DX:AX = *addr
	// all in one instruction
	LOCK
	CMPXCHG8B	0(BP)
	JNZ	swaploop

	// success
	// return DX:AX
	MOVL	AX, ret_lo+12(FP)
	MOVL	DX, ret_hi+16(FP)
	RET

// uint64 atomicload64(uint64 volatile* addr);
TEXT ·Load64(SB), NOSPLIT, $0-12
	NO_LOCAL_POINTERS
	MOVL	ptr+0(FP), AX
	TESTL	$7, AX
	JZ	2(PC)
	CALL	·panicUnaligned(SB)
	MOVQ	(AX), M0
	MOVQ	M0, ret+4(FP)
	EMMS
	RET

// void ·Store64(uint64 volatile* addr, uint64 v);
TEXT ·Store64(SB), NOSPLIT, $0-12
	NO_LOCAL_POINTERS
	MOVL	ptr+0(FP), AX
	TESTL	$7, AX
	JZ	2(PC)
	CALL	·panicUnaligned(SB)
	// MOVQ and EMMS were introduced on the Pentium MMX.
	MOVQ	val+4(FP), M0
	MOVQ	M0, (AX)
	EMMS
	// This is essentially a no-op, but it provides required memory fencing.
	// It can be replaced with MFENCE, but MFENCE was introduced only on the Pentium4 (SSE2).
	XORL	AX, AX
	LOCK
	XADDL	AX, (SP)
	RET

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-20
	MOVL	ptr+0(FP), BP
	// DI:SI = v
	MOVL	val_lo+4(FP), SI
	MOVL	val_hi+8(FP), DI
	// DX:AX = *addr
	MOVL	0(BP), AX
	MOVL	4(BP), DX
casloop:
	// CX:BX = DX:AX (*addr) & DI:SI (mask)
	MOVL	AX, BX
	MOVL	DX, CX
	ANDL	SI, BX
	ANDL	DI, CX
	LOCK
	CMPXCHG8B	0(BP)
	JNZ casloop
	MOVL	AX, ret_lo+12(FP)
	MOVL	DX, ret_hi+16(FP)
	RET

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-20
	MOVL	ptr+0(FP), BP
	// DI:SI = v
	MOVL	val_lo+4(FP), SI
	MOVL	val_hi+8(FP), DI
	// DX:AX = *addr
	MOVL	0(BP), AX
	MOVL	4(BP), DX
casloop:
	// CX:BX = DX:AX (*addr) | DI:SI (mask)
	MOVL	AX, BX
	MOVL	DX, CX
	ORL	SI, BX
	ORL	DI, CX
	LOCK
	CMPXCHG8B	0(BP)
	JNZ casloop
	MOVL	AX, ret_lo+12(FP)
	MOVL	DX, ret_hi+16(FP)
	RET
