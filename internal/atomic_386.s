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
TEXT ·SwapIfLessUint64(SB), NOSPLIT, $8-37
	NO_LOCAL_POINTERS
	MOVL	addr+0(FP), BP
	TESTL	$7, BP
	JZ	aligned
	CALL	·panicUnaligned(SB)
aligned:
	MOVL	new_lo+4(FP), BX
	MOVL	new_hi+8(FP), CX
	MOVL	BX, DI
	XORL	sign_lo+20(FP), DI
	ANDL	mask_lo+12(FP), DI
	MOVL	DI, 0(SP)
	MOVL	CX, DI
	XORL	sign_hi+24(FP), DI
	ANDL	mask_hi+16(FP), DI
	MOVL	DI, 4(SP)
	// The no-swap path must also observe an atomic 64-bit value.
	// Use an aligned MMX load rather than two MOVLs.
	MOVQ	(BP), M0
	MOVQ	M0, old+28(FP)
	EMMS
	MOVL	old_lo+28(FP), AX
	MOVL	old_hi+32(FP), DX
retry:
	MOVL	DX, DI
	XORL	sign_hi+24(FP), DI
	ANDL	mask_hi+16(FP), DI
	CMPL	DI, 4(SP)
	JHI	replace
	JCS	unchanged
	MOVL	AX, DI
	XORL	sign_lo+20(FP), DI
	ANDL	mask_lo+12(FP), DI
	CMPL	DI, 0(SP)
	JLS	unchanged
replace:
	// Failed CAS supplies a coherent new DX:AX for the next comparison.
	LOCK
	CMPXCHG8B	(BP)
	JNE	retry
	MOVB	$1, swapped+36(FP)
	JMP	done
unchanged:
	MOVB	$0, swapped+36(FP)
done:
	MOVL	AX, old_lo+28(FP)
	MOVL	DX, old_hi+32(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-37
	MOVL	sign_lo+20(FP), AX
	MOVL	sign_hi+24(FP), DX
	XORL	mask_lo+12(FP), AX
	XORL	mask_hi+16(FP), DX
	MOVL	AX, sign_lo+20(FP)
	MOVL	DX, sign_hi+24(FP)
	JMP	·SwapIfLessUint64(SB)
