// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (mips64 || mips64le) && !race

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
	MOVV	addr+0(FP), R1
	MOVV	new+8(FP), R2
	MOVV	mask+16(FP), R3
	MOVV	sign+24(FP), R4
	XOR	R4, R2, R5
	AND	R3, R5
	SYNC
retry:
	LLV	(R1), R6
	XOR	R4, R6, R7
	AND	R3, R7
	SGTU	R7, R5, R8
	BEQ	R8, unchanged
	MOVV	R2, R8
	SCV	R8, (R1)
	BEQ	R8, retry
	MOVV	$1, R8
	JMP	done
unchanged:
	MOVV	$0, R8
done:
	SYNC
	MOVV	R6, old+32(FP)
	MOVB	R8, swapped+40(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-41
	MOVV	sign+24(FP), R1
	MOVV	mask+16(FP), R2
	XOR	R2, R1
	MOVV	R1, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)
