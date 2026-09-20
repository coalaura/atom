// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

// TODO(neelance): implement with actual atomic operations as soon as threads are available
// See https://github.com/WebAssembly/design/issues/1073

package internal

//go:nosplit
//go:noinline
func Load64(ptr *uint64) uint64 {
	return *ptr
}

//go:nosplit
//go:noinline
func Xadd64(ptr *uint64, delta int64) uint64 {
	new := *ptr + uint64(delta)
	*ptr = new
	return new
}

//go:nosplit
//go:noinline
func Xchg64(ptr *uint64, new uint64) uint64 {
	old := *ptr
	*ptr = new
	return old
}

//go:nosplit
//go:noinline
func Cas64(ptr *uint64, old, new uint64) bool {
	if *ptr == old {
		*ptr = new
		return true
	}
	return false
}

//go:nosplit
//go:noinline
func Store64(ptr *uint64, val uint64) {
	*ptr = val
}
