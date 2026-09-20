package internal_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/coalaura/atom/internal"
)

const (
	uint64Pattern    = uint64(0xaaaaaaaa55555555)
	uint64Complement = ^uint64Pattern
	uint64Iterations = 10000
)

type alignedUint64 struct {
	_     [0]atomic.Uint64
	value uint64
}

type uint64OperationCase struct {
	initial uint64
	operand uint64
}

func TestUint64Operations(t *testing.T) {
	cases := []uint64OperationCase{
		{initial: 0, operand: 1 << 63},
		{initial: ^uint64(0), operand: 1},
		{initial: 1 << 32, operand: ^uint64(0)},
		{initial: uint64Pattern, operand: uint64Complement},
	}

	for _, testCase := range cases {
		var (
			actual   alignedUint64
			expected atomic.Uint64
		)

		internal.StoreUint64(&actual.value, testCase.initial)
		expected.Store(testCase.initial)
		checkUint64Value(t, &actual, &expected)

		got := internal.AddUint64(&actual.value, testCase.operand)
		want := expected.Add(testCase.operand)

		if got != want {
			t.Fatalf("AddUint64(%#x, %#x) = %#x, want %#x", testCase.initial, testCase.operand, got, want)
		}

		checkUint64Value(t, &actual, &expected)

		got = internal.SwapUint64(&actual.value, testCase.initial)
		want = expected.Swap(testCase.initial)

		if got != want {
			t.Fatalf("SwapUint64 returned %#x, want %#x", got, want)
		}

		checkUint64Value(t, &actual, &expected)

		swapped := internal.CompareAndSwapUint64(&actual.value, testCase.initial^1, testCase.operand)
		if swapped {
			t.Fatal("CompareAndSwapUint64 succeeded with a mismatched old value")
		}

		checkUint64Value(t, &actual, &expected)

		swapped = internal.CompareAndSwapUint64(&actual.value, testCase.initial, testCase.operand)
		if !swapped {
			t.Fatal("CompareAndSwapUint64 failed with the current value")
		}

		expected.Store(testCase.operand)
		checkUint64Value(t, &actual, &expected)

		got = internal.AndUint64(&actual.value, testCase.initial)
		want = expected.And(testCase.initial)

		if got != want {
			t.Fatalf("AndUint64 returned %#x, want %#x", got, want)
		}

		checkUint64Value(t, &actual, &expected)

		got = internal.OrUint64(&actual.value, testCase.operand)
		want = expected.Or(testCase.operand)

		if got != want {
			t.Fatalf("OrUint64 returned %#x, want %#x", got, want)
		}

		checkUint64Value(t, &actual, &expected)
	}
}

func TestUint64LoadStoreConcurrent(t *testing.T) {
	var (
		value     alignedUint64
		waitGroup sync.WaitGroup
	)

	internal.StoreUint64(&value.value, uint64Pattern)
	waitGroup.Add(2)

	for range 2 {
		go func() {
			defer waitGroup.Done()

			for range uint64Iterations {
				internal.StoreUint64(&value.value, uint64Pattern)
				internal.StoreUint64(&value.value, uint64Complement)
			}
		}()
	}

	for range uint64Iterations {
		loaded := internal.LoadUint64(&value.value)
		if loaded != uint64Pattern && loaded != uint64Complement {
			t.Errorf("LoadUint64 observed a torn value: %#x", loaded)

			break
		}
	}

	waitGroup.Wait()
}

func TestUint64BitwiseConcurrent(t *testing.T) {
	var (
		value     alignedUint64
		waitGroup sync.WaitGroup
	)

	waitGroup.Add(64)

	for bit := range 64 {
		go func() {
			defer waitGroup.Done()

			mask := uint64(1) << bit
			old := internal.OrUint64(&value.value, mask)

			if old&mask != 0 {
				t.Errorf("OrUint64 returned a value with bit %d already set: %#x", bit, old)
			}
		}()
	}

	waitGroup.Wait()

	loaded := internal.LoadUint64(&value.value)
	if loaded != ^uint64(0) {
		t.Fatalf("OrUint64 lost updates: got %#x", loaded)
	}

	waitGroup.Add(64)

	for bit := range 64 {
		go func() {
			defer waitGroup.Done()

			mask := uint64(1) << bit
			old := internal.AndUint64(&value.value, ^mask)

			if old&mask == 0 {
				t.Errorf("AndUint64 returned a value with bit %d already cleared: %#x", bit, old)
			}
		}()
	}

	waitGroup.Wait()

	loaded = internal.LoadUint64(&value.value)
	if loaded != 0 {
		t.Fatalf("AndUint64 lost updates: got %#x", loaded)
	}
}

func checkUint64Value(t *testing.T, actual *alignedUint64, expected *atomic.Uint64) {
	t.Helper()

	got := internal.LoadUint64(&actual.value)
	want := expected.Load()

	if got != want {
		t.Fatalf("LoadUint64() = %#x, want %#x", got, want)
	}
}
