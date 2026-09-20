// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

package internal

// SwapIfLessInteger compares at width bytes with the specified signedness.
// A compact type description keeps the caller inlineable and lets assembly
// specialize full-width comparisons without constructing ordering keys.
// The caller derives success by comparing new against the returned old value.
//
//go:noescape
func SwapIfLessInteger(addr *uint64, new uint64, width uint8, signed bool) (old uint64)

// SwapIfGreaterInteger uses the same type description and swaps only if greater.
//
//go:noescape
func SwapIfGreaterInteger(addr *uint64, new uint64, width uint8, signed bool) (old uint64)
