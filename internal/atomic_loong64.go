// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build loong64 && !race

package internal

import (
	"github.com/coalaura/atom/internal/cpu"
	"unsafe"
)

const (
	offsetLOONG64HasLAMCAS     = unsafe.Offsetof(cpu.Loong64.HasLAMCAS)
	offsetLoong64HasDBAR_HINTS = unsafe.Offsetof(cpu.Loong64.HasDBAR_HINTS)
)

//go:noescape
func Xadd64(ptr *uint64, delta int64) uint64

//go:noescape
func Xchg64(ptr *uint64, new uint64) uint64

//go:noescape
func Load64(ptr *uint64) uint64

//go:noescape
func And64(ptr *uint64, val uint64) uint64

//go:noescape
func Or64(ptr *uint64, val uint64) uint64

//go:noescape
func Cas64(ptr *uint64, old, new uint64) bool

//go:noescape
func Store64(ptr *uint64, val uint64)
