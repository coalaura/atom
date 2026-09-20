// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "textflag.h"
#include "funcdata.h"

// func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Atomically:
//	old = *addr
//	swapped = (new^sign)&mask < (old^sign)&mask
//	if swapped {
//		*addr = new
//	}
//	return old, swapped
//
// Go's wasm execution is single-threaded. There are no calls or scheduling
// points between load and store.
TEXT ·SwapIfLessUint64(SB),NOSPLIT,$16-41
	NO_LOCAL_POINTERS
	MOVD	addr+0(FP), R0
	// Linear-memory address zero is valid to wasm, but is a nil pointer in Go.
	Get	R0
	I64Eqz
	If
		// Load64 raises a recoverable nil-pointer panic.
		MOVD	R0, 0(SP)
		CALLNORESUME ·Load64(SB)
	End
	MOVD	new+8(FP), R1
	MOVD	mask+16(FP), R2
	MOVD	sign+24(FP), R3
	MOVD	0(R0), R4
	Get	R1
	Get	R3
	I64Xor
	Get	R2
	I64And
	Get	R4
	Get	R3
	I64Xor
	Get	R2
	I64And
	I64LtU
	If
		MOVD	R1, 0(R0)
		MOVB	$1, swapped+40(FP)
	Else
		MOVB	$0, swapped+40(FP)
	End
	MOVD	R4, old+32(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB),NOSPLIT,$0-41
	MOVD	sign+24(FP), R0
	MOVD	mask+16(FP), R1
	Get	R0
	Get	R1
	I64Xor
	Set	R0
	MOVD	R0, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)
