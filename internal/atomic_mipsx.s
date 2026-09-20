// Copyright 2016 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (mips || mipsle) && !race

#include "textflag.h"
#include "funcdata.h"

TEXT ·spinLock(SB),NOSPLIT,$0-4
	MOVW	state+0(FP), R1
	MOVW	$1, R2
	SYNC
try_lock:
	MOVW	R2, R3
check_again:
	LL	(R1), R4
	BNE	R4, check_again
	SC	R3, (R1)
	BEQ	R3, try_lock
	SYNC
	RET

TEXT ·spinUnlock(SB),NOSPLIT,$0-4
	MOVW	state+0(FP), R1
	SYNC
	MOVW	R0, (R1)
	SYNC
	RET

// Both the argument words and the pointed-to value follow native endianness.
#ifdef GOARCH_mips
#define NEW_LO new_lo+8(FP)
#define NEW_HI new_hi+4(FP)
#define MASK_LO mask_lo+16(FP)
#define MASK_HI mask_hi+12(FP)
#define SIGN_LO sign_lo+24(FP)
#define SIGN_HI sign_hi+20(FP)
#define OLD_LO old_lo+32(FP)
#define OLD_HI old_hi+28(FP)
#define VALUE_LO 4(R1)
#define VALUE_HI 0(R1)
#else
#define NEW_LO new_lo+4(FP)
#define NEW_HI new_hi+8(FP)
#define MASK_LO mask_lo+12(FP)
#define MASK_HI mask_hi+16(FP)
#define SIGN_LO sign_lo+20(FP)
#define SIGN_HI sign_hi+24(FP)
#define OLD_LO old_lo+28(FP)
#define OLD_HI old_hi+32(FP)
#define VALUE_LO 0(R1)
#define VALUE_HI 4(R1)
#endif

// func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Atomically:
//	old = *addr
//	swapped = (new^sign)&mask < (old^sign)&mask
//	if swapped {
//		*addr = new
//	}
//	return old, swapped
TEXT ·SwapIfLessUint64(SB),NOSPLIT,$8-37
	NO_LOCAL_POINTERS
	MOVW	addr+0(FP), R1
	MOVW	R1, 4(R29)
	// Use the same lock, alignment check, and barriers as the other 64-bit operations.
	JAL	·lockAndCheck(SB)
	MOVW	addr+0(FP), R1
	MOVW	NEW_LO, R2
	MOVW	NEW_HI, R3
	MOVW	VALUE_LO, R4
	MOVW	VALUE_HI, R5
	MOVW	MASK_LO, R6
	MOVW	MASK_HI, R7
	MOVW	SIGN_LO, R8
	MOVW	SIGN_HI, R9
	XOR	R9, R5, R10
	AND	R7, R10
	XOR	R9, R3, R11
	AND	R7, R11
	SGTU	R10, R11, R12
	BNE	R12, replace
	BNE	R10, R11, unchanged
	XOR	R8, R4, R10
	AND	R6, R10
	XOR	R8, R2, R11
	AND	R6, R11
	SGTU	R10, R11, R12
	BEQ	R12, unchanged
replace:
	MOVW	R2, VALUE_LO
	MOVW	R3, VALUE_HI
	MOVW	$1, R12
	JMP	done
unchanged:
	MOVW	$0, R12
done:
	MOVW	R4, OLD_LO
	MOVW	R5, OLD_HI
	MOVB	R12, swapped+36(FP)
	JAL	·unlock(SB)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB),NOSPLIT,$-4-37
	MOVW	SIGN_LO, R1
	MOVW	SIGN_HI, R2
	MOVW	MASK_LO, R3
	MOVW	MASK_HI, R4
	XOR	R3, R1
	XOR	R4, R2
	MOVW	R1, SIGN_LO
	MOVW	R2, SIGN_HI
	JMP	·SwapIfLessUint64(SB)
