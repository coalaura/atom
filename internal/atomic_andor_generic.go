// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (arm || wasm) && !race

package internal

//go:nosplit
func And64(ptr *uint64, val uint64) uint64 {
	for {
		old := *ptr
		if Cas64(ptr, old, old&val) {
			return old
		}
	}
}

//go:nosplit
func Or64(ptr *uint64, val uint64) uint64 {
	for {
		old := *ptr
		if Cas64(ptr, old, old|val) {
			return old
		}
	}
}
