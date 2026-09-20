// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "textflag.h"

// func SwapIfLessInteger(addr *uint64, new uint64, width uint8, signed bool) (old uint64)
// Atomically, with comparisons at the specified width and signedness:
//	old = *addr
//	if new < old {
//		*addr = new
//	}
//	return old
TEXT ·SwapIfLessInteger(SB), NOSPLIT, $0-32
	MOVQ	addr+0(FP), BX
	MOVQ	new+8(FP), DX
	MOVQ	(BX), AX
	CMPB	width+16(FP), $8
	JNE	narrow
	CMPB	signed+17(FP), $0
	JE	unsigned
signed:
	CMPQ	DX, AX
	JGE	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	signed
	MOVQ	AX, old+24(FP)
	RET
unsigned:
	CMPQ	DX, AX
	JCC	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	unsigned
	MOVQ	AX, old+24(FP)
	RET
narrow:
	// Shift away overflow bits and move the type's sign bit to bit 63.
	// AX and DX retain the full old and new values for the CAS.
	MOVBQZX	width+16(FP), CX
	SHLQ	$3, CX
	NEGQ	CX
	ADDQ	$64, CX
	MOVQ	DX, SI
	SHLQ	CX, SI
	CMPB	signed+17(FP), $0
	JE	narrowunsigned
narrowsigned:
	MOVQ	AX, DI
	SHLQ	CX, DI
	CMPQ	SI, DI
	JGE	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	narrowsigned
	MOVQ	AX, old+24(FP)
	RET
narrowunsigned:
	MOVQ	AX, DI
	SHLQ	CX, DI
	CMPQ	SI, DI
	JCC	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	narrowunsigned
unchanged:
	MOVQ	AX, old+24(FP)
	RET

// func SwapIfGreaterInteger(addr *uint64, new uint64, width uint8, signed bool) (old uint64)
// Atomically, with comparisons at the specified width and signedness:
//	old = *addr
//	if new > old {
//		*addr = new
//	}
//	return old
TEXT ·SwapIfGreaterInteger(SB), NOSPLIT, $0-32
	MOVQ	addr+0(FP), BX
	MOVQ	new+8(FP), DX
	MOVQ	(BX), AX
	CMPB	width+16(FP), $8
	JNE	narrow
	CMPB	signed+17(FP), $0
	JE	unsigned
signed:
	CMPQ	DX, AX
	JLE	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	signed
	MOVQ	AX, old+24(FP)
	RET
unsigned:
	CMPQ	DX, AX
	JLS	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	unsigned
	MOVQ	AX, old+24(FP)
	RET
narrow:
	// Shift away overflow bits without changing the full-width CAS operands.
	MOVBQZX	width+16(FP), CX
	SHLQ	$3, CX
	NEGQ	CX
	ADDQ	$64, CX
	MOVQ	DX, SI
	SHLQ	CX, SI
	CMPB	signed+17(FP), $0
	JE	narrowunsigned
narrowsigned:
	MOVQ	AX, DI
	SHLQ	CX, DI
	CMPQ	SI, DI
	JLE	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	narrowsigned
	MOVQ	AX, old+24(FP)
	RET
narrowunsigned:
	MOVQ	AX, DI
	SHLQ	CX, DI
	CMPQ	SI, DI
	JLS	unchanged
	LOCK
	CMPXCHGQ DX, (BX)
	JNE	narrowunsigned
unchanged:
	MOVQ	AX, old+24(FP)
	RET
