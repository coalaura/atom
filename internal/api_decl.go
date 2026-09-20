// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 && !race

package internal

// SwapIfLessUint64 compares unsigned ordering keys (value ^ sign) & mask.
// mask selects the bits belonging to the integer type; sign is its sign bit
// for signed types, or zero for unsigned types. The full original bits are
// used for CAS and returned, even when Add has overflowed a narrower type.
// Equal keys do not cause a store, including when the unused upper bits differ.
//
//go:noescape
func SwapIfLessUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)

// SwapIfGreaterUint64 uses the same keys and swaps only when new is greater.
// Assembly complements sign within mask to reverse the unsigned ordering,
// then tail-calls SwapIfLessUint64 to share its atomic retry loop.
//
//go:noescape
func SwapIfGreaterUint64(addr *uint64, new, mask, sign uint64) (old uint64, swapped bool)
