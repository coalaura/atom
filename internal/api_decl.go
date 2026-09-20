// Copyright 2023 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

package internal

//go:noescape
func SwapUint64(addr *uint64, new uint64) (old uint64)

//go:noescape
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool)

//go:noescape
func AddUint64(addr *uint64, delta uint64) (new uint64)

//go:noescape
func AndUint64(addr *uint64, mask uint64) (old uint64)

//go:noescape
func OrUint64(addr *uint64, mask uint64) (old uint64)

//go:noescape
func LoadUint64(addr *uint64) (val uint64)

//go:noescape
func StoreUint64(addr *uint64, val uint64)
