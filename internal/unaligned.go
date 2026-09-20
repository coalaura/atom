// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (386 || arm) && !race

package internal

func panicUnaligned() {
	panic("unaligned 64-bit atomic operation")
}
