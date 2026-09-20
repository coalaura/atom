package atom_test

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/coalaura/atom"
)

const (
	intTestGoroutines = 8
	intTestOps        = 1000
)

type testInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type namedInt16 int16

type intAlignmentContainer struct {
	_     byte
	value atom.Int[int64]
}

func TestIntTypes(t *testing.T) {
	t.Run("int", testInt[int])
	t.Run("int8", testInt[int8])
	t.Run("int16", testInt[int16])
	t.Run("int32", testInt[int32])
	t.Run("int64", testInt[int64])
	t.Run("uint", testInt[uint])
	t.Run("uint8", testInt[uint8])
	t.Run("uint16", testInt[uint16])
	t.Run("uint32", testInt[uint32])
	t.Run("uint64", testInt[uint64])
	t.Run("uintptr", testInt[uintptr])
	t.Run("named", testInt[namedInt16])
}

func TestIntSignedOverflow(t *testing.T) {
	var value atom.Int[int8]

	value.Store(127)

	got := value.Add(1)
	if got != -128 {
		t.Fatalf("Add(1) = %v, want -128", got)
	}

	got = value.Load()
	if got != -128 {
		t.Fatalf("Load() = %v, want -128", got)
	}

	value.Store(-128)

	swapped := value.CompareAndSwap(-128, -127)
	if !swapped {
		t.Fatal("CompareAndSwap(-128, -127) failed, want true")
	}

	value.Store(-1)

	got = value.And(0x7f)
	if got != -1 {
		t.Fatalf("And(0x7f) = %v, want -1", got)
	}

	got = value.Load()
	if got != 0x7f {
		t.Fatalf("Load() after And = %v, want 127", got)
	}

	value.Store(-128)

	got = value.Or(1)
	if got != -128 {
		t.Fatalf("Or(1) = %v, want -128", got)
	}

	got = value.Load()
	if got != -127 {
		t.Fatalf("Load() after Or = %v, want -127", got)
	}
}

func TestIntUnsignedOverflow(t *testing.T) {
	var value atom.Int[uint8]

	value.Store(255)

	got := value.Add(1)
	if got != 0 {
		t.Fatalf("Add(1) = %v, want 0", got)
	}

	got = value.Load()
	if got != 0 {
		t.Fatalf("Load() = %v, want 0", got)
	}
}

func TestIntConcurrent(t *testing.T) {
	var (
		value      atom.Int[int64]
		waitGroup  sync.WaitGroup
		goroutines [intTestGoroutines]struct{}
		operations [intTestOps]struct{}
	)

	waitGroup.Add(intTestGoroutines)

	for range goroutines {
		go func() {
			defer waitGroup.Done()

			for range operations {
				value.Add(1)
			}
		}()
	}

	waitGroup.Wait()

	want := int64(intTestGoroutines * intTestOps)
	got := value.Load()

	if got != want {
		t.Fatalf("Load() = %v, want %v", got, want)
	}
}

func TestIntAlignment(t *testing.T) {
	var container intAlignmentContainer

	address := uintptr(unsafe.Pointer(&container.value))
	if address%8 != 0 {
		t.Fatalf("Int address = %#x, want 64-bit alignment", address)
	}
}

func testInt[T testInteger](t *testing.T) {
	var value atom.Int[T]

	got := value.Load()
	if got != 0 {
		t.Fatalf("zero Load() = %v, want 0", got)
	}

	value.Store(10)

	got = value.Load()
	if got != 10 {
		t.Fatalf("Load() = %v, want 10", got)
	}

	got = value.Swap(20)
	if got != 10 {
		t.Fatalf("Swap(20) = %v, want 10", got)
	}

	swapped := value.CompareAndSwap(10, 30)
	if swapped {
		t.Fatal("CompareAndSwap(10, 30) succeeded, want false")
	}

	swapped = value.CompareAndSwap(20, 30)
	if !swapped {
		t.Fatal("CompareAndSwap(20, 30) failed, want true")
	}

	got = value.Add(5)
	if got != 35 {
		t.Fatalf("Add(5) = %v, want 35", got)
	}

	value.Store(0b1110)

	got = value.And(0b1011)
	if got != 0b1110 {
		t.Fatalf("And(0b1011) = %b, want 1110", got)
	}

	got = value.Load()
	if got != 0b1010 {
		t.Fatalf("Load() after And = %b, want 1010", got)
	}

	got = value.Or(0b0101)
	if got != 0b1010 {
		t.Fatalf("Or(0b0101) = %b, want 1010", got)
	}

	got = value.Load()
	if got != 0b1111 {
		t.Fatalf("Load() after Or = %b, want 1111", got)
	}
}
