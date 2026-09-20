// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm && !race

// Package cpu preserves the internal/cpu layout needed to calculate assembly
// offsets. Assembly reads the Go runtime's feature variables, not this copy.
package cpu

// CacheLinePad is used to pad structs to avoid false sharing.
type CacheLinePad struct{ _ [CacheLinePadSize]byte }

// The booleans in ARM contain the correspondingly named cpu feature bit.
// The struct is padded to avoid false sharing.
var ARM struct {
	_            CacheLinePad
	HasVFPv4     bool
	HasIDIVA     bool
	HasV7Atomics bool
	_            CacheLinePad
}
