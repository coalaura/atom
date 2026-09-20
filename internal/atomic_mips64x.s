// Copyright 2015 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (mips64 || mips64le) && !race

#include "textflag.h"

#define SYNC	WORD $0xf

// func	Cas64(ptr *uint64, old, new uint64) bool
// Atomically:
//	if *ptr == old {
//		*ptr = new
//		return true
//	} else {
//		return false
//	}
TEXT ·Cas64(SB), NOSPLIT, $0-25
	MOVV	ptr+0(FP), R1
	MOVV	old+8(FP), R2
	MOVV	new+16(FP), R5
	SYNC
cas64_again:
	MOVV	R5, R3
	LLV	(R1), R4
	BNE	R2, R4, cas64_fail
	SCV	R3, (R1)
	BEQ	R3, cas64_again
	MOVV	$1, R1
	MOVB	R1, ret+24(FP)
	SYNC
	RET
cas64_fail:
	MOVV	$0, R1
	JMP	-4(PC)

// uint64 Xadd64(uint64 volatile *ptr, int64 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R2
	MOVV	delta+8(FP), R3
	SYNC
	LLV	(R2), R1
	ADDVU	R1, R3, R4
	MOVV	R4, R1
	SCV	R4, (R2)
	BEQ	R4, -4(PC)
	MOVV	R1, ret+16(FP)
	SYNC
	RET

// uint64 Xchg64(ptr *uint64, new uint64)
// Atomically:
//	old := *ptr;
//	*ptr = new;
//	return old;
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R2
	MOVV	new+8(FP), R5

	SYNC
	MOVV	R5, R3
	LLV	(R2), R1
	SCV	R3, (R2)
	BEQ	R3, -3(PC)
	MOVV	R1, ret+16(FP)
	SYNC
	RET

TEXT ·Store64(SB), NOSPLIT, $0-16
	MOVV	ptr+0(FP), R1
	MOVV	val+8(FP), R2
	SYNC
	MOVV	R2, 0(R1)
	SYNC
	RET

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R1
	MOVV	val+8(FP), R2

	SYNC
	LLV	(R1), R3
	OR	R2, R3, R4
	SCV	R4, (R1)
	BEQ	R4, -3(PC)
	SYNC
	MOVV	R3, ret+16(FP)
	RET

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R1
	MOVV	val+8(FP), R2

	SYNC
	LLV	(R1), R3
	AND	R2, R3, R4
	SCV	R4, (R1)
	BEQ	R4, -3(PC)
	SYNC
	MOVV	R3, ret+16(FP)
	RET

// uint64 ·Load64(uint64 volatile* ptr)
TEXT ·Load64(SB),NOSPLIT|NOFRAME,$0-16
	MOVV	ptr+0(FP), R1
	SYNC
	MOVV	0(R1), R1
	SYNC
	MOVV	R1, ret+8(FP)
	RET

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
