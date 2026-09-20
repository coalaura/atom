// Copyright 2014 The Go Authors. All rights reserved.
// Copyright 2026 coalaura (github.com/coalaura). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atom

import (
	"sync/atomic"
	"unsafe"
)

// A Value provides an atomic load and store of a value of type T.
// The zero value for a Value returns the zero value of T from [Value.Load].
// Once [Value.Store] has been called, a Value must not be copied.
//
// A Value must not be copied after first use.
type Value[T any] struct {
	v any
}

// efaceWords is any's internal representation.
type efaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

// Load returns the value set by the most recent Store.
// It returns the zero value of T if there has been no call to Store for this Value.
func (v *Value[T]) Load() (val T) {
	vp := (*efaceWords)(unsafe.Pointer(v))
	typ := atomic.LoadPointer(&vp.typ)
	if typ == nil || typ == unsafe.Pointer(&firstStoreInProgress) {
		// First store not yet completed.
		return
	}
	data := atomic.LoadPointer(&vp.data)
	// Reconstruct via any so the eface layout is independent of T.
	var i any
	vlp := (*efaceWords)(unsafe.Pointer(&i))
	vlp.typ = typ
	vlp.data = data
	return i.(T)
}

var firstStoreInProgress byte

// Store sets the value of the [Value] v to val.
// All calls to Store for a given Value must use values of the same concrete type.
// Store of an inconsistent type panics, as does Store of a nil interface value.
func (v *Value[T]) Store(val T) {
	// Box into any so we can inspect/store the eface words. A true nil
	// interface panics; typed nils and concrete zeros do not.
	var iface any = val
	if iface == nil {
		panic("sync/atomic: store of nil value into Value")
	}
	vp := (*efaceWords)(unsafe.Pointer(v))
	vlp := (*efaceWords)(unsafe.Pointer(&iface))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion.
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&firstStoreInProgress)) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, vlp.data)
			atomic.StorePointer(&vp.typ, vlp.typ)
			runtime_procUnpin()
			return
		}
		if typ == unsafe.Pointer(&firstStoreInProgress) {
			// First store in progress. Wait.
			// Since we disable preemption around the first store,
			// we can wait with active spinning.
			continue
		}
		// First store completed. Check type and overwrite data.
		if typ != vlp.typ {
			panic("sync/atomic: store of inconsistently typed value into Value")
		}
		atomic.StorePointer(&vp.data, vlp.data)
		return
	}
}

// Swap stores new into Value and returns the previous value. It returns the
// zero value of T if the Value is empty.
//
// All calls to Swap for a given Value must use values of the same concrete
// type. Swap of an inconsistent type panics, as does Swap of a nil interface value.
func (v *Value[T]) Swap(new T) (old T) {
	var iface any = new
	if iface == nil {
		panic("sync/atomic: swap of nil value into Value")
	}
	vp := (*efaceWords)(unsafe.Pointer(v))
	np := (*efaceWords)(unsafe.Pointer(&iface))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion.
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&firstStoreInProgress)) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, np.data)
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return
		}
		if typ == unsafe.Pointer(&firstStoreInProgress) {
			// First store in progress. Wait.
			// Since we disable preemption around the first store,
			// we can wait with active spinning.
			continue
		}
		// First store completed. Check type and overwrite data.
		if typ != np.typ {
			panic("sync/atomic: swap of inconsistently typed value into Value")
		}
		var oldIface any
		op := (*efaceWords)(unsafe.Pointer(&oldIface))
		op.typ, op.data = np.typ, atomic.SwapPointer(&vp.data, np.data)
		return oldIface.(T)
	}
}

// SwapIfFunc stores new into Value when shouldSwap returns true for the
// currently stored value. It returns the value observed by shouldSwap and
// whether the swap succeeded. For an empty Value, shouldSwap receives the zero
// value of T.
//
// If another goroutine changes the Value while shouldSwap is running,
// SwapIfFunc retries with the latest value. Therefore, shouldSwap may be called
// multiple times and must be safe for concurrent use.
//
// All calls to SwapIfFunc for a given Value must use values of the same
// concrete type. SwapIfFunc with an inconsistent type panics, as does
// SwapIfFunc with a nil interface value for new.
func (v *Value[T]) SwapIfFunc(new T, shouldSwap func(old T) bool) (old T, swapped bool) {
	var newIface any = new
	if newIface == nil {
		panic("sync/atomic: swap if func of nil value into Value")
	}

	vp := (*efaceWords)(unsafe.Pointer(v))
	np := (*efaceWords)(unsafe.Pointer(&newIface))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			if !shouldSwap(old) {
				return old, false
			}

			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion.
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&firstStoreInProgress)) {
				runtime_procUnpin()
				continue
			}

			// Complete first store.
			atomic.StorePointer(&vp.data, np.data)
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return old, true
		}
		if typ == unsafe.Pointer(&firstStoreInProgress) {
			// First store in progress. Wait.
			// Since we disable preemption around the first store,
			// we can wait with active spinning.
			continue
		}
		if typ != np.typ {
			panic("sync/atomic: swap if func of inconsistently typed value into Value")
		}

		data := atomic.LoadPointer(&vp.data)
		var oldIface any
		op := (*efaceWords)(unsafe.Pointer(&oldIface))
		op.typ, op.data = typ, data
		old = oldIface.(T)
		if !shouldSwap(old) {
			return old, false
		}
		if atomic.CompareAndSwapPointer(&vp.data, data, np.data) {
			return old, true
		}
	}
}

// CompareAndSwap executes the compare-and-swap operation for the [Value].
//
// All calls to CompareAndSwap for a given Value must use values of the same
// concrete type. CompareAndSwap of an inconsistent type panics, as does
// CompareAndSwap(old, new) where new is a nil interface value.
func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	var newIface any = new
	if newIface == nil {
		panic("sync/atomic: compare and swap of nil value into Value")
	}
	vp := (*efaceWords)(unsafe.Pointer(v))
	np := (*efaceWords)(unsafe.Pointer(&newIface))
	var oldIface any = old
	op := (*efaceWords)(unsafe.Pointer(&oldIface))
	if op.typ != nil && np.typ != op.typ {
		panic("sync/atomic: compare and swap of inconsistently typed values")
	}
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			// Empty Value: only succeeds when old is a true nil interface
			// (same rule as the original sync/atomic.Value).
			if oldIface != nil {
				return false
			}
			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion.
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&firstStoreInProgress)) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, np.data)
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return true
		}
		if typ == unsafe.Pointer(&firstStoreInProgress) {
			// First store in progress. Wait.
			// Since we disable preemption around the first store,
			// we can wait with active spinning.
			continue
		}
		// First store completed. Check type and overwrite data.
		if typ != np.typ {
			panic("sync/atomic: compare and swap of inconsistently typed value into Value")
		}
		// Compare old and current via runtime equality check.
		// This allows value types to be compared, something
		// not offered by the package functions.
		// CompareAndSwapPointer below only ensures vp.data
		// has not changed since LoadPointer.
		data := atomic.LoadPointer(&vp.data)
		var i any
		(*efaceWords)(unsafe.Pointer(&i)).typ = typ
		(*efaceWords)(unsafe.Pointer(&i)).data = data
		if i != oldIface {
			return false
		}
		return atomic.CompareAndSwapPointer(&vp.data, data, np.data)
	}
}

//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
