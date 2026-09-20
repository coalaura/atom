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
