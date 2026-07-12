package atom_test

import (
	"testing"

	"github.com/coalaura/atom"
)

const (
	testGoroutines = 8
	testOps        = 1000
)

func TestValueLoadEmpty(t *testing.T) {
	var v atom.Value[int]

	if got := v.Load(); got != 0 {
		t.Fatalf("Load empty = %v, want 0", got)
	}

	var s atom.Value[string]

	if got := s.Load(); got != "" {
		t.Fatalf("Load empty = %q, want \"\"", got)
	}
}

func TestValueStoreLoad(t *testing.T) {
	var v atom.Value[int]

	v.Store(42)

	if got := v.Load(); got != 42 {
		t.Fatalf("Load = %v, want 42", got)
	}

	v.Store(7)

	if got := v.Load(); got != 7 {
		t.Fatalf("Load = %v, want 7", got)
	}
}

func TestValueSwap(t *testing.T) {
	var v atom.Value[string]

	if got := v.Swap("a"); got != "" {
		t.Fatalf("Swap empty = %q, want \"\"", got)
	}

	if got := v.Load(); got != "a" {
		t.Fatalf("Load = %q, want \"a\"", got)
	}

	if got := v.Swap("b"); got != "a" {
		t.Fatalf("Swap = %q, want \"a\"", got)
	}

	if got := v.Load(); got != "b" {
		t.Fatalf("Load = %q, want \"b\"", got)
	}
}

func TestValueCompareAndSwap(t *testing.T) {
	var v atom.Value[int]

	if v.CompareAndSwap(0, 1) {
		t.Fatal("CAS on empty with old=0 succeeded, want false")
	}

	if got := v.Load(); got != 0 {
		t.Fatalf("Load = %v, want 0", got)
	}

	v.Store(1)

	if !v.CompareAndSwap(1, 2) {
		t.Fatal("CAS(1, 2) failed, want success")
	}

	if got := v.Load(); got != 2 {
		t.Fatalf("Load = %v, want 2", got)
	}

	if v.CompareAndSwap(1, 3) {
		t.Fatal("CAS(1, 3) succeeded, want false")
	}

	if got := v.Load(); got != 2 {
		t.Fatalf("Load = %v, want 2", got)
	}
}

func TestValueCompareAndSwapInterfaceNil(t *testing.T) {
	var v atom.Value[any]

	if !v.CompareAndSwap(nil, "x") {
		t.Fatal("CAS(nil, \"x\") on empty failed, want success")
	}

	if got := v.Load(); got != "x" {
		t.Fatalf("Load = %v, want \"x\"", got)
	}

	if v.CompareAndSwap(nil, "y") {
		t.Fatal("CAS(nil, \"y\") after store succeeded, want false")
	}
}

func TestValueStoreNilPanics(t *testing.T) {
	var v atom.Value[any]

	defer func() {
		if recover() == nil {
			t.Fatal("Store(nil) did not panic")
		}
	}()

	v.Store(nil)
}

func TestValueSwapNilPanics(t *testing.T) {
	var v atom.Value[any]

	defer func() {
		if recover() == nil {
			t.Fatal("Swap(nil) did not panic")
		}
	}()

	v.Swap(nil)
}

func TestValueCompareAndSwapNilPanics(t *testing.T) {
	var v atom.Value[any]

	v.Store("x")

	defer func() {
		if recover() == nil {
			t.Fatal("CompareAndSwap(_, nil) did not panic")
		}
	}()

	v.CompareAndSwap("x", nil)
}

func TestValueInconsistentTypePanics(t *testing.T) {
	var v atom.Value[any]

	v.Store(1)

	defer func() {
		if recover() == nil {
			t.Fatal("Store of inconsistent type did not panic")
		}
	}()

	v.Store("x")
}

func TestValueConcurrent(t *testing.T) {
	var v atom.Value[int]
	v.Store(0)

	done := make(chan struct{}, testGoroutines)
	for g := 0; g < testGoroutines; g++ {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			for i := 0; i < testOps; i++ {
				v.Store(id)
				_ = v.Load()
				_ = v.Swap(id)
				v.CompareAndSwap(id, id+1)
			}
		}(g)
	}
	for g := 0; g < testGoroutines; g++ {
		<-done
	}
	_ = v.Load()
}
