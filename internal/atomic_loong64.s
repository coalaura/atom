// Copyright 2022 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "go_asm.h"
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
	MOVV	ptr+0(FP), R4
	MOVV	old+8(FP), R5
	MOVV	new+16(FP), R6

	MOVBU	internal∕cpu·Loong64+const_offsetLOONG64HasLAMCAS(SB), R8
	BEQ	R8, ll_sc_64
	MOVV	R5, R7  // backup old value
	AMCASDBV	R6, (R4), R5
	BNE	R7, R5, cas64_fail0
	MOVV	$1, R4
	MOVB	R4, ret+24(FP)
	RET
cas64_fail0:
	MOVB	R0, ret+24(FP)
	RET

ll_sc_64:
	// Implemented using the ll-sc instruction pair
	DBAR	$0x14
cas64_again:
	MOVV	R6, R7
	LLV	(R4), R8
	BNE	R5, R8, cas64_fail1
	SCV	R7, (R4)
	BEQ	R7, cas64_again
	MOVV	$1, R4
	MOVB	R4, ret+24(FP)
	DBAR	$0x12
	RET
cas64_fail1:
	MOVV	$0, R4
	JMP	-4(PC)

// func Xadd64(ptr *uint64, delta int64) uint64
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R4
	MOVV	delta+8(FP), R5
	AMADDDBV	R5, (R4), R6
	ADDV	R6, R5, R4
	MOVV	R4, ret+16(FP)
	RET

// func Xchg64(ptr *uint64, new uint64) uint64
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R4
	MOVV	new+8(FP), R5
	AMSWAPDBV	R5, (R4), R6
	MOVV	R6, ret+16(FP)
	RET

TEXT ·Store64(SB), NOSPLIT, $0-16
	MOVV	ptr+0(FP), R4
	MOVV	val+8(FP), R5
	MOVBU	internal∕cpu·Loong64+const_offsetLoong64HasDBAR_HINTS(SB), R6
	BEQ	R6, _variant_
	// StoreRelease barrier
	DBAR	$0x12
	MOVV	R5, 0(R4)
	DBAR	$0x18
	RET
_variant_:
	AMSWAPDBV	R5, (R4), R0
	RET

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R4
	MOVV	val+8(FP), R5
	AMORDBV	R5, (R4), R6
	MOVV	R6, ret+16(FP)
	RET

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOVV	ptr+0(FP), R4
	MOVV	val+8(FP), R5
	AMANDDBV	R5, (R4), R6
	MOVV	R6, ret+16(FP)
	RET

// uint64 internal∕runtime∕atomic·Load64(uint64 volatile* ptr)
TEXT ·Load64(SB),NOSPLIT|NOFRAME,$0-16
	MOVV	ptr+0(FP), R19
	MOVV	0(R19), R19
	DBAR	$0x14
	MOVV	R19, ret+8(FP)
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
	MOVV	addr+0(FP), R4
	MOVV	new+8(FP), R5
	MOVV	mask+16(FP), R6
	MOVV	sign+24(FP), R7
	XOR	R7, R5, R8
	AND	R6, R8
	// Match the barriers of the baseline Cas64 LL/SC implementation.
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
