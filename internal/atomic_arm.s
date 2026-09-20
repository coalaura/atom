// Copyright 2015 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "go_asm.h"
#include "textflag.h"
#include "funcdata.h"

// The following functions all panic if their address argument isn't
// 8-byte aligned. Since we're calling back into Go code to do this,
// we have to cooperate with stack unwinding. In the normal case, the
// functions tail-call into the appropriate implementation, which
// means they must not open a frame. Hence, when they go down the
// panic path, at that point they push the LR to create a real frame
// (they don't need to pop it because panic won't return; however, we
// do need to set the SP delta back).

// Check if R1 is 8-byte aligned, panic if not.
// Clobbers R2.
#define CHECK_ALIGN \
	AND.S	$7, R1, R2 \
	BEQ 	4(PC) \
	MOVW.W	R14, -4(R13) /* prepare a real frame */ \
	BL	·panicUnaligned(SB) \
	ADD	$4, R13 /* compensate SP delta */

// func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Atomically:
//	old = *addr
//	swapped = (new^sign)&mask < (old^sign)&mask
//	if swapped {
//		*addr = new
//	}
//	return old, swapped
TEXT ·SwapIfLessUint64(SB),NOSPLIT,$-4-37
	NO_LOCAL_POINTERS
	MOVW	addr+0(FP), R1
	CHECK_ALIGN
#ifndef GOARM_7
	MOVB	internal∕cpu·ARM+const_offsetARMHasV7Atomics(SB), R11
	CMP	$1, R11
	BEQ	2(PC)
	JMP	·goSwapIfLess64(SB)
#endif
	JMP	swapIfLessNative<>(SB)

TEXT swapIfLessNative<>(SB),NOSPLIT,$0-37
	// R1 is the checked address. R2:R3 and R4:R5 are even register pairs
	// for STREXD and LDREXD; R10 is reserved for g and must not be used.
	MOVW	new_lo+4(FP), R2
	MOVW	new_hi+8(FP), R3
	MOVW	mask_lo+12(FP), R6
	MOVW	mask_hi+16(FP), R7
	MOVW	sign_lo+20(FP), R8
	MOVW	sign_hi+24(FP), R9
retry:
	LDREXD	(R1), R4
	EOR	R9, R3, R11
	AND	R7, R11
	EOR	R9, R5, R0
	AND	R7, R0
	CMP	R0, R11
	BLO	replace
	BHI	unchanged
	EOR	R8, R2, R11
	AND	R6, R11
	EOR	R8, R4, R0
	AND	R6, R0
	CMP	R0, R11
	BHS	unchanged
replace:
	DMB	MB_ISHST
	STREXD	R2, (R1), R0
	CMP	$0, R0
	BNE	retry
	MOVW	$1, R0
	B	done
unchanged:
	MOVW	$0, R0
done:
	// The read-only outcome needs the acquire barrier too.
	DMB	MB_ISH
	MOVW	R4, old_lo+28(FP)
	MOVW	R5, old_hi+32(FP)
	MOVB	R0, swapped+36(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB),NOSPLIT,$-4-37
	MOVW	sign_lo+20(FP), R0
	MOVW	sign_hi+24(FP), R1
	MOVW	mask_lo+12(FP), R2
	MOVW	mask_hi+16(FP), R3
	EOR	R2, R0
	EOR	R3, R1
	MOVW	R0, sign_lo+20(FP)
	MOVW	R1, sign_hi+24(FP)
	JMP	·SwapIfLessUint64(SB)
