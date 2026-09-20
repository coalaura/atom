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

	got := v.Load()
	if got != 0 {
		t.Fatalf("Load empty = %v, want 0", got)
	}

	var s atom.Value[string]

	gotString := s.Load()
	if gotString != "" {
		t.Fatalf("Load empty = %q, want \"\"", gotString)
	}
}

func TestValueStoreLoad(t *testing.T) {
	var v atom.Value[int]

	v.Store(42)

	got := v.Load()
	if got != 42 {
		t.Fatalf("Load = %v, want 42", got)
	}

	v.Store(7)

	got = v.Load()
	if got != 7 {
		t.Fatalf("Load = %v, want 7", got)
	}
}

func TestValueSwap(t *testing.T) {
	var v atom.Value[string]

	got := v.Swap("a")
	if got != "" {
		t.Fatalf("Swap empty = %q, want \"\"", got)
	}

	got = v.Load()
	if got != "a" {
		t.Fatalf("Load = %q, want \"a\"", got)
	}

	got = v.Swap("b")
	if got != "a" {
		t.Fatalf("Swap = %q, want \"a\"", got)
	}

	got = v.Load()
	if got != "b" {
		t.Fatalf("Load = %q, want \"b\"", got)
	}
}

func TestValueSwapIfFunc(t *testing.T) {
	var value atom.Value[int]

	old, swapped := value.SwapIfFunc(1, func(old int) bool {
		return old == 1
	})

	if old != 0 || swapped {
		t.Fatalf("SwapIfFunc on empty = (%v, %v), want (0, false)", old, swapped)
	}

	old, swapped = value.SwapIfFunc(1, func(old int) bool {
		return old == 0
	})

	if old != 0 || !swapped {
		t.Fatalf("SwapIfFunc on empty = (%v, %v), want (0, true)", old, swapped)
	}

	old, swapped = value.SwapIfFunc(2, func(old int) bool {
		return old == 0
	})

	if old != 1 || swapped {
		t.Fatalf("SwapIfFunc false = (%v, %v), want (1, false)", old, swapped)
	}

	old, swapped = value.SwapIfFunc(2, func(old int) bool {
		return old == 1
	})

	if old != 1 || !swapped {
		t.Fatalf("SwapIfFunc true = (%v, %v), want (1, true)", old, swapped)
	}

	got := value.Load()
	if got != 2 {
		t.Fatalf("Load = %v, want 2", got)
	}
}

func TestValueSwapIfFuncNonComparable(t *testing.T) {
	var value atom.Value[[]int]

	value.Store([]int{1, 2})

	old, swapped := value.SwapIfFunc([]int{3, 4}, func(old []int) bool {
		return len(old) == 2 && old[0] == 1 && old[1] == 2
	})

	if len(old) != 2 || old[0] != 1 || old[1] != 2 || !swapped {
		t.Fatalf("SwapIfFunc = (%v, %v), want ([1 2], true)", old, swapped)
	}

	got := value.Load()
	if len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("Load = %v, want [3 4]", got)
	}
}

func TestValueSwapIfFuncInconsistentTypePanics(t *testing.T) {
	var value atom.Value[any]

	value.Store(1)

	defer func() {
		if recover() == nil {
			t.Fatal("SwapIfFunc of inconsistent type did not panic")
		}
	}()

	value.SwapIfFunc("x", func(any) bool {
		t.Fatal("predicate called for an inconsistent type")

		return true
	})
}

func TestValueSwapIfFuncConcurrent(t *testing.T) {
	var value atom.Value[int]

	value.Store(0)

	results := make(chan bool, testGoroutines)
	goroutines := [testGoroutines]struct{}{}

	for goroutine := range goroutines {
		go func(candidate int) {
			_, swapped := value.SwapIfFunc(candidate, func(old int) bool {
				return old == 0
			})

			results <- swapped
		}(goroutine + 1)
	}

	swaps := 0

	for range goroutines {
		if <-results {
			swaps++
		}
	}

	if swaps != 1 {
		t.Fatalf("successful swaps = %v, want 1", swaps)
	}
}

func TestValueCompareAndSwap(t *testing.T) {
	var v atom.Value[int]

	swapped := v.CompareAndSwap(0, 1)
	if swapped {
		t.Fatal("CAS on empty with old=0 succeeded, want false")
	}

	got := v.Load()
	if got != 0 {
		t.Fatalf("Load = %v, want 0", got)
	}

	v.Store(1)

	swapped = v.CompareAndSwap(1, 2)
	if !swapped {
		t.Fatal("CAS(1, 2) failed, want success")
	}

	got = v.Load()
	if got != 2 {
		t.Fatalf("Load = %v, want 2", got)
	}

	swapped = v.CompareAndSwap(1, 3)
	if swapped {
		t.Fatal("CAS(1, 3) succeeded, want false")
	}

	got = v.Load()
	if got != 2 {
		t.Fatalf("Load = %v, want 2", got)
	}
}

func TestValueCompareAndSwapInterfaceNil(t *testing.T) {
	var v atom.Value[any]

	swapped := v.CompareAndSwap(nil, "x")
	if !swapped {
		t.Fatal("CAS(nil, \"x\") on empty failed, want success")
	}

	got := v.Load()
	if got != "x" {
		t.Fatalf("Load = %v, want \"x\"", got)
	}

	swapped = v.CompareAndSwap(nil, "y")
	if swapped {
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

	var (
		goroutines [testGoroutines]struct{}
		operations [testOps]struct{}
	)

	for goroutine := range goroutines {
		go func(id int) {
			defer func() {
				done <- struct{}{}
			}()

			for range operations {
				v.Store(id)
				_ = v.Load()
				_ = v.Swap(id)
				v.CompareAndSwap(id, id+1)
			}
		}(goroutine)
	}

	for range goroutines {
		<-done
	}

	_ = v.Load()
}
