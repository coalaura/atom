// Copyright 2015 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

// Note: some of these functions are semantically inlined
// by the compiler (in src/cmd/compile/internal/gc/ssa.go).

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
	MOVQ	ptr+0(FP), BX
	MOVQ	old+8(FP), AX
	MOVQ	new+16(FP), CX
	LOCK
	CMPXCHGQ	CX, 0(BX)
	SETEQ	ret+24(FP)
	RET

// uint64 Xadd64(uint64 volatile *val, int64 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	delta+8(FP), AX
	MOVQ	AX, CX
	LOCK
	XADDQ	AX, 0(BX)
	ADDQ	CX, AX
	MOVQ	AX, ret+16(FP)
	RET

// uint64 Xchg64(ptr *uint64, new uint64)
// Atomically:
//	old := *ptr;
//	*ptr = new;
//	return old;
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	new+8(FP), AX
	XCHGQ	AX, 0(BX)
	MOVQ	AX, ret+16(FP)
	RET

TEXT ·Store64(SB), NOSPLIT, $0-16
	MOVQ	ptr+0(FP), BX
	MOVQ	val+8(FP), AX
	XCHGQ	AX, 0(BX)
	RET

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	val+8(FP), CX
casloop:
	MOVQ 	CX, DX
	MOVQ	(BX), AX
	ORQ	AX, DX
	LOCK
	CMPXCHGQ	DX, (BX)
	JNZ casloop
	MOVQ 	AX, ret+16(FP)
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
	MOVQ	addr+0(FP), BX
	MOVQ	new+8(FP), CX
	MOVQ	mask+16(FP), SI
	MOVQ	sign+24(FP), DI
	MOVQ	CX, R8
	XORQ	DI, R8
	ANDQ	SI, R8
	MOVQ	(BX), AX
retry:
	MOVQ	AX, DX
	XORQ	DI, DX
	ANDQ	SI, DX
	CMPQ	R8, DX
	JCC	unchanged
	// On failure CMPXCHGQ refreshes AX with the full current value.
	LOCK
	CMPXCHGQ	CX, (BX)
	JNE	retry
	MOVB	$1, swapped+40(FP)
	MOVQ	AX, old+32(FP)
	RET
unchanged:
	MOVB	$0, swapped+40(FP)
	MOVQ	AX, old+32(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-41
	MOVQ	sign+24(FP), AX
	XORQ	mask+16(FP), AX
	MOVQ	AX, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOVQ	ptr+0(FP), BX
	MOVQ	val+8(FP), CX
casloop:
	MOVQ 	CX, DX
	MOVQ	(BX), AX
	ANDQ	AX, DX
	LOCK
	CMPXCHGQ	DX, (BX)
	JNZ casloop
	MOVQ 	AX, ret+16(FP)
	RET
