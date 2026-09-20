// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "textflag.h"

// func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Atomically:
//	old = *addr
//	swapped = (new^sign)&mask < (old^sign)&mask
//	if swapped {
//		*addr = new
//	}
//	return old, swapped
TEXT ·SwapIfLessUint64(SB), NOSPLIT, $0-41
	MOVD	addr+0(FP), R0
	MOVD	new+8(FP), R1
	MOVD	mask+16(FP), R2
	MOVD	sign+24(FP), R3
	EOR	R3, R1, R4
	AND	R2, R4
retry:
	LDAXR	(R0), R5
	EOR	R3, R5, R6
	AND	R2, R6
	CMP	R6, R4
	BHS	unchanged
	STLXR	R1, (R0), R7
	CBNZ	R7, retry
	MOVD	$1, R7
	MOVB	R7, swapped+40(FP)
	MOVD	R5, old+32(FP)
	RET
unchanged:
	CLREX
	MOVB	ZR, swapped+40(FP)
	MOVD	R5, old+32(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-41
	MOVD	sign+24(FP), R0
	MOVD	mask+16(FP), R1
	EOR	R1, R0
	MOVD	R0, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)
