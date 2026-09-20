// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "textflag.h"

// func Store64(ptr *uint64, val uint64)
TEXT ·Store64(SB), NOSPLIT, $0
	MOVD	ptr+0(FP), R2
	MOVD	val+8(FP), R3
	MOVD	R3, 0(R2)
	SYNC
	RET

// func Cas64(ptr *uint64, old, new uint64) bool
// Atomically:
//	if *ptr == old {
//		*ptr = new
//		return 1
//	} else {
//		return 0
//	}
TEXT ·Cas64(SB), NOSPLIT, $0-25
	MOVD	ptr+0(FP), R3
	MOVD	old+8(FP), R4
	MOVD	new+16(FP), R5
	CSG	R4, R5, 0(R3)    //  if (R4 == 0(R3)) then 0(R3)= R5
	BNE	cas64_fail
	MOVB	$1, ret+24(FP)
	RET
cas64_fail:
	MOVB	$0, ret+24(FP)
	RET

// func Xadd64(ptr *uint64, delta int64) uint64
TEXT ·Xadd64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R4
	MOVD	delta+8(FP), R5
	MOVD	(R4), R3
repeat:
	ADD	R5, R3, R6
	CSG	R3, R6, (R4) // if R3==(R4) then (R4)=R6 else R3=(R4)
	BNE	repeat
	MOVD	R6, ret+16(FP)
	RET

// func Xchg64(ptr *uint64, new uint64) uint64
TEXT ·Xchg64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R4
	MOVD	new+8(FP), R3
	MOVD	(R4), R6
repeat:
	CSG	R6, R3, (R4) // if R6==(R4) then (R4)=R3 else R6=(R4)
	BNE	repeat
	MOVD	R6, ret+16(FP)
	RET

// func Or64(addr *uint64, v uint64) old uint64
TEXT ·Or64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R4
	MOVD	val+8(FP), R5
	MOVD	(R4), R3
repeat:
	OR	R5, R3, R6
	CSG	R3, R6, (R4) // if R3==(R4) then (R4)=R6 else R3=(R4)
	BNE	repeat
	MOVD	R3, ret+16(FP)
	RET

// func And64(addr *uint64, v uint64) old uint64
TEXT ·And64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R4
	MOVD	val+8(FP), R5
	MOVD	(R4), R3
repeat:
	AND	R5, R3, R6
	CSG	R3, R6, (R4) // if R3==(R4) then (R4)=R6 else R3=(R4)
	BNE	repeat
	MOVD	R3, ret+16(FP)
	RET
