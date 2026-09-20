// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (arm || arm64 || loong64 || mips || mipsle) && !race

// Package cpu preserves the internal/cpu layouts needed to calculate assembly
// offsets. Assembly reads the Go runtime's feature variables, not these copies.
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

// The booleans in ARM64 contain the correspondingly named cpu feature bit.
// The struct is padded to avoid false sharing.
var ARM64 struct {
	_          CacheLinePad
	HasAES     bool
	HasPMULL   bool
	HasSHA1    bool
	HasSHA2    bool
	HasSHA512  bool
	HasSHA3    bool
	HasCRC32   bool
	HasATOMICS bool
	HasCPUID   bool
	HasDIT     bool
	HasSB      bool
	IsNeoverse bool
	_          CacheLinePad
}

// The booleans in Loong64 contain the correspondingly named cpu feature bit.
// The struct is padded to avoid false sharing.
var Loong64 struct {
	_              CacheLinePad
	HasLSX         bool // support 128-bit vector extension
	HasLASX        bool // support 256-bit vector extension
	HasCRC32       bool // support CRC instruction
	HasLAMCAS      bool // support AMCAS[_DB].{B/H/W/D}
	HasLAM_BH      bool // support AM{SWAP/ADD}[_DB].{B/H} instruction
	HasLLACQ_SCREL bool // support LLACQ.{W/D}, SCREL.{W/D} instruction
	HasSCQ         bool // support SC.Q instruction
	HasDBAR_HINTS  bool // supports finer-grained DBAR hints
	_              CacheLinePad
}
