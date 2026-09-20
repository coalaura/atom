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
	MOVV	addr+0(FP), R4
	MOVV	new+8(FP), R5
	MOVV	mask+16(FP), R6
	MOVV	sign+24(FP), R7
	XOR	R7, R5, R8
	AND	R6, R8
	// Match the barriers of the stdlib's baseline Cas64 LL/SC implementation.
	DBAR	$0x14
retry:
	LLV	(R4), R9
	XOR	R7, R9, R10
	AND	R6, R10
	SGTU	R10, R8, R11
	BEQ	R11, unchanged
	MOVV	R5, R11
	SCV	R11, (R4)
	BEQ	R11, retry
	MOVV	$1, R11
	JMP	done
unchanged:
	MOVV	$0, R11
done:
	DBAR	$0x12
	MOVV	R9, old+32(FP)
	MOVB	R11, swapped+40(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-41
	MOVV	sign+24(FP), R4
	MOVV	mask+16(FP), R5
	XOR	R5, R4
	MOVV	R4, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)
