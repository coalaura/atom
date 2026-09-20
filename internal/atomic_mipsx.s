// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (mips || mipsle) && !race

#include "textflag.h"

// Argument words follow native endianness.
#ifdef GOARCH_mips
#define MASK_LO mask_lo+16(FP)
#define MASK_HI mask_hi+12(FP)
#define SIGN_LO sign_lo+24(FP)
#define SIGN_HI sign_hi+20(FP)
#else
#define MASK_LO mask_lo+12(FP)
#define MASK_HI mask_hi+16(FP)
#define SIGN_LO sign_lo+20(FP)
#define SIGN_HI sign_hi+24(FP)
#endif

// func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Atomically:
//	old = *addr
//	swapped = (new^sign)&mask < (old^sign)&mask
//	if swapped {
//		*addr = new
//	}
//	return old, swapped
TEXT ·SwapIfLessUint64(SB),NOSPLIT,$-4-37
	// Use stdlib atomics so all accesses participate in the same locking protocol.
	JMP	·goSwapIfLess64(SB)

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
