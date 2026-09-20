// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (ppc64 || ppc64le) && !race

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
	MOVD	addr+0(FP), R3
	MOVD	new+8(FP), R4
	MOVD	mask+16(FP), R5
	MOVD	sign+24(FP), R6
	XOR	R6, R4, R7
	AND	R5, R7
	// Also order the read-only outcome, as in the stdlib's Load64.
	SYNC
retry:
	LDAR	(R3), R8
	XOR	R6, R8, R9
	AND	R5, R9
	CMPU	R7, R9
	BGE	unchanged
	STDCCC	R4, (R3)
	BNE	retry
	MOVD	$1, R10
	JMP	done
unchanged:
	MOVD	$0, R10
done:
	LWSYNC
	MOVD	R8, old+32(FP)
	MOVB	R10, swapped+40(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-41
	MOVD	sign+24(FP), R3
	MOVD	mask+16(FP), R4
	XOR	R4, R3
	MOVD	R3, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)
