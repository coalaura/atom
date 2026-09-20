// Copyright 2014 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "go_asm.h"
#include "textflag.h"

// uint64 ·Load64(uint64 volatile* addr)
TEXT ·Load64(SB),NOSPLIT,$0-16
	MOVD	ptr+0(FP), R0
	LDAR	(R0), R0
	MOVD	R0, ret+8(FP)
	RET

TEXT ·Store64(SB), NOSPLIT, $0-16
	MOVD	ptr+0(FP), R0
	MOVD	val+8(FP), R1
	STLR	R1, (R0)
	RET

// uint64 Xchg64(ptr *uint64, new uint64)
// Atomically:
//	old := *ptr;
//	*ptr = new;
//	return old;
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R0
	MOVD	new+8(FP), R1
#ifndef GOARM64_LSE
	MOVBU	internal∕cpu·ARM64+const_offsetARM64HasATOMICS(SB), R4
	CBZ 	R4, load_store_loop
#endif
	SWPALD	R1, (R0), R2
	MOVD	R2, ret+16(FP)
	RET
#ifndef GOARM64_LSE
load_store_loop:
	LDAXR	(R0), R2
	STLXR	R1, (R0), R3
	CBNZ	R3, load_store_loop
	MOVD	R2, ret+16(FP)
	RET
#endif

// func Cas64(ptr *uint64, old, new uint64) bool
// Atomically:
//	if *ptr == old {
//		*ptr = new
//		return true
//	} else {
//		return false
//	}
TEXT ·Cas64(SB), NOSPLIT, $0-25
	MOVD	ptr+0(FP), R0
	MOVD	old+8(FP), R1
	MOVD	new+16(FP), R2
#ifndef GOARM64_LSE
	MOVBU	internal∕cpu·ARM64+const_offsetARM64HasATOMICS(SB), R4
	CBZ 	R4, load_store_loop
#endif
	MOVD	R1, R3
	CASALD	R3, (R0), R2
	CMP 	R1, R3
	CSET	EQ, R0
	MOVB	R0, ret+24(FP)
	RET
#ifndef GOARM64_LSE
load_store_loop:
	LDAXR	(R0), R3
	CMP	R1, R3
	BNE	ok
	STLXR	R2, (R0), R3
	CBNZ	R3, load_store_loop
ok:
	CSET	EQ, R0
	MOVB	R0, ret+24(FP)
	RET
#endif

// uint64 Xadd64(uint64 volatile *ptr, int64 delta)
// Atomically:
//      *val += delta;
//      return *val;
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R0
	MOVD	delta+8(FP), R1
#ifndef GOARM64_LSE
	MOVBU	internal∕cpu·ARM64+const_offsetARM64HasATOMICS(SB), R4
	CBZ 	R4, load_store_loop
#endif
	LDADDALD	R1, (R0), R2
	ADD 	R1, R2
	MOVD	R2, ret+16(FP)
	RET
#ifndef GOARM64_LSE
load_store_loop:
	LDAXR	(R0), R2
	ADD	R2, R1, R2
	STLXR	R2, (R0), R3
	CBNZ	R3, load_store_loop
	MOVD	R2, ret+16(FP)
	RET
#endif

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R0
	MOVD	val+8(FP), R1
#ifndef GOARM64_LSE
	MOVBU	internal∕cpu·ARM64+const_offsetARM64HasATOMICS(SB), R4
	CBZ 	R4, load_store_loop
#endif
	LDORALD	R1, (R0), R2
	MOVD	R2, ret+16(FP)
	RET
#ifndef GOARM64_LSE
load_store_loop:
	LDAXR	(R0), R2
	ORR	R1, R2, R3
	STLXR	R3, (R0), R4
	CBNZ	R4, load_store_loop
	MOVD 	R2, ret+16(FP)
	RET
#endif

// func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Atomically:
//	old = *addr
//	swapped = (new^sign)&mask < (old^sign)&mask
//	if swapped {
//		*addr = new
//	}
//	return old, swapped
TEXT ·SwapIfLessUint64(SB), NOSPLIT, $0-41
	MOVD	addr+0(FP), R0
	MOVD	new+8(FP), R1
	MOVD	mask+16(FP), R2
	MOVD	sign+24(FP), R3
	EOR	R3, R1, R4
	AND	R2, R4
retry:
	LDAXR	(R0), R5
	EOR	R3, R5, R6
	AND	R2, R6
	CMP	R6, R4
	BHS	unchanged
	STLXR	R1, (R0), R7
	CBNZ	R7, retry
	MOVD	$1, R7
	MOVB	R7, swapped+40(FP)
	MOVD	R5, old+32(FP)
	RET
unchanged:
	CLREX
	MOVB	ZR, swapped+40(FP)
	MOVD	R5, old+32(FP)
	RET

// func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
// Complementing the masked key bits reverses their unsigned order.
TEXT ·SwapIfGreaterUint64(SB), NOSPLIT, $0-41
	MOVD	sign+24(FP), R0
	MOVD	mask+16(FP), R1
	EOR	R1, R0
	MOVD	R0, sign+24(FP)
	JMP	·SwapIfLessUint64(SB)

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R0
	MOVD	val+8(FP), R1
#ifndef GOARM64_LSE
	MOVBU	internal∕cpu·ARM64+const_offsetARM64HasATOMICS(SB), R4
	CBZ 	R4, load_store_loop
#endif
	MVN 	R1, R2
	LDCLRALD	R2, (R0), R3
	MOVD	R3, ret+16(FP)
	RET
#ifndef GOARM64_LSE
load_store_loop:
	LDAXR	(R0), R2
	AND	R1, R2, R3
	STLXR	R3, (R0), R4
	CBNZ	R4, load_store_loop
	MOVD 	R2, ret+16(FP)
	RET
#endif
