// Copyright 2011 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

#include "textflag.h"

TEXT ·SwapUint64(SB),NOSPLIT,$0
	JMP	·Xchg64(SB)

TEXT ·CompareAndSwapUint64(SB),NOSPLIT,$0
	JMP	·Cas64(SB)

TEXT ·AddUint64(SB),NOSPLIT,$0
	JMP	·Xadd64(SB)

TEXT ·LoadUint64(SB),NOSPLIT,$0
	JMP	·Load64(SB)

TEXT ·StoreUint64(SB),NOSPLIT,$0
	JMP	·Store64(SB)

TEXT ·AndUint64(SB),NOSPLIT,$0
	JMP	·And64(SB)

TEXT ·OrUint64(SB),NOSPLIT,$0
	JMP	·Or64(SB)
