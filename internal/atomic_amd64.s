// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

// Note: some of these functions are semantically inlined
// by the compiler (in src/cmd/compile/internal/gc/ssa.go).

#include "textflag.h"

// func Cas64(ptr *uint64, old, new uint64) bool
// Atomically:
//	if *ptr == old {
//		*ptr = new
//		return true
//	} else {
//		return false
//	}
TEXT ·Cas64(SB), NOSPLIT, $0-25
	MOVQ	ptr+0(FP), BX
	MOVQ	old+8(FP), AX
	MOVQ	new+16(FP), CX
	LOCK
	CMPXCHGQ	CX, 0(BX)
	SETEQ	ret+24(FP)
	RET

// uint64 Xadd64(uint64 volatile *val, int64 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	delta+8(FP), AX
	MOVQ	AX, CX
	LOCK
	XADDQ	AX, 0(BX)
	ADDQ	CX, AX
	MOVQ	AX, ret+16(FP)
	RET

// uint64 Xchg64(ptr *uint64, new uint64)
// Atomically:
//	old := *ptr;
//	*ptr = new;
//	return old;
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	new+8(FP), AX
	XCHGQ	AX, 0(BX)
	MOVQ	AX, ret+16(FP)
	RET

TEXT ·Store64(SB), NOSPLIT, $0-16
	MOVQ	ptr+0(FP), BX
	MOVQ	val+8(FP), AX
	XCHGQ	AX, 0(BX)
	RET

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	val+8(FP), CX
casloop:
	MOVQ 	CX, DX
	MOVQ	(BX), AX
	ORQ	AX, DX
	LOCK
	CMPXCHGQ	DX, (BX)
	JNZ casloop
	MOVQ 	AX, ret+16(FP)
	RET

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	val+8(FP), CX
casloop:
	MOVQ 	CX, DX
	MOVQ	(BX), AX
	ANDQ	AX, DX
	LOCK
	CMPXCHGQ	DX, (BX)
	JNZ casloop
	MOVQ 	AX, ret+16(FP)
	RET
