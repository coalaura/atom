// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm && !race

package internal

import (
	"github.com/coalaura/atom/internal/cpu"
	"unsafe"
)

const (
	offsetARMHasV7Atomics = unsafe.Offsetof(cpu.ARM.HasV7Atomics)
)
