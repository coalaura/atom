// Copyright 2014 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build race

package internal

import "sync/atomic"

func SwapUint64(addr *uint64, new uint64) (old uint64) {
	return atomic.SwapUint64(addr, new)
}

func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool) {
	return atomic.CompareAndSwapUint64(addr, old, new)
}

func AddUint64(addr *uint64, delta uint64) (new uint64) {
	return atomic.AddUint64(addr, delta)
}

func AndUint64(addr *uint64, mask uint64) (old uint64) {
	return atomic.AndUint64(addr, mask)
}

func OrUint64(addr *uint64, mask uint64) (old uint64) {
	return atomic.OrUint64(addr, mask)
}

func LoadUint64(addr *uint64) (val uint64) {
	return atomic.LoadUint64(addr)
}

func StoreUint64(addr *uint64, val uint64) {
	atomic.StoreUint64(addr, val)
}
