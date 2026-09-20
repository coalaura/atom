package atom_test

import (
	"math/rand/v2"
	"sync"
	"testing"
	"unsafe"

	"github.com/coalaura/atom"
)

type namedUint64 uint64

type intConditionalSwapCase[T testInteger] struct {
	name    string
	initial T
	new     T
	less    bool
	greater bool
}

func TestIntConditionalSwaps(t *testing.T) {
	t.Run("int", testIntConditionalSwaps[int])
	t.Run("int8", testIntConditionalSwaps[int8])
	t.Run("int16", testIntConditionalSwaps[int16])
	t.Run("int32", testIntConditionalSwaps[int32])
	t.Run("int64", testIntConditionalSwaps[int64])
	t.Run("uint", testIntConditionalSwaps[uint])
	t.Run("uint8", testIntConditionalSwaps[uint8])
	t.Run("uint16", testIntConditionalSwaps[uint16])
	t.Run("uint32", testIntConditionalSwaps[uint32])
	t.Run("uint64", testIntConditionalSwaps[uint64])
	t.Run("uintptr", testIntConditionalSwaps[uintptr])
	t.Run("namedInt16", testIntConditionalSwaps[namedInt16])
	t.Run("namedUint64", testIntConditionalSwaps[namedUint64])
}

func TestIntConditionalSwapsConcurrent(t *testing.T) {
	t.Run("less", func(t *testing.T) {
		testIntConditionalSwapsConcurrent(t, false)
	})

	t.Run("greater", func(t *testing.T) {
		testIntConditionalSwapsConcurrent(t, true)
	})
}

func TestIntConditionalSwapsSignedness(t *testing.T) {
	var (
		signed   atom.Int[int64]
		unsigned atom.Int[uint64]
	)

	negative := int64(-34)
	signed.Store(12)
	unsigned.Store(12)

	oldSigned, swapped := signed.SwapIfLess(negative)
	if oldSigned != 12 || !swapped {
		t.Fatalf("signed SwapIfLess(-34) = (%v, %v), want (12, true)", oldSigned, swapped)
	}

	gotSigned := signed.Load()
	if gotSigned != negative {
		t.Fatalf("signed Load = %v, want -34", gotSigned)
	}

	// Converting the variable preserves the same bits but gives them unsigned ordering.
	oldUnsigned, swapped := unsigned.SwapIfLess(uint64(negative))
	if oldUnsigned != 12 || swapped {
		t.Fatalf("unsigned SwapIfLess = (%v, %v), want (12, false)", oldUnsigned, swapped)
	}

	gotUnsigned := unsigned.Load()
	if gotUnsigned != 12 {
		t.Fatalf("unsigned Load = %v, want 12", gotUnsigned)
	}

	oldSigned, swapped = signed.SwapIfGreater(12)
	if oldSigned != negative || !swapped {
		t.Fatalf("signed SwapIfGreater(12) = (%v, %v), want (-34, true)", oldSigned, swapped)
	}

	oldUnsigned, swapped = unsigned.SwapIfGreater(uint64(negative))
	if oldUnsigned != 12 || !swapped {
		t.Fatalf("unsigned SwapIfGreater = (%v, %v), want (12, true)", oldUnsigned, swapped)
	}

	gotSigned = signed.Load()
	if gotSigned != 12 {
		t.Fatalf("signed Load after greater = %v, want 12", gotSigned)
	}

	gotUnsigned = unsigned.Load()
	if gotUnsigned != uint64(negative) {
		t.Fatalf("unsigned Load after greater = %v, want %v", gotUnsigned, uint64(negative))
	}
}

func TestIntConditionalSwapsNil(t *testing.T) {
	t.Run("less", func(t *testing.T) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				t.Fatal("SwapIfLess on a nil receiver did not panic")
			}
		}()

		var value *atom.Int[int64]

		value.SwapIfLess(1)
	})

	t.Run("greater", func(t *testing.T) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				t.Fatal("SwapIfGreater on a nil receiver did not panic")
			}
		}()

		var value *atom.Int[int64]

		value.SwapIfGreater(1)
	})
}

func testIntConditionalSwaps[T testInteger](t *testing.T) {
	var minimum T

	maximum := ^T(0)
	if maximum < 0 {
		minimum = T(1) << (unsafe.Sizeof(minimum)*8 - 1)
		maximum = ^minimum
	}

	cases := []intConditionalSwapCase[T]{
		{name: "zero", initial: 0, new: 0},
		{name: "equal minimum", initial: minimum, new: minimum},
		{name: "equal maximum", initial: maximum, new: maximum},
		{name: "minimum to maximum", initial: minimum, new: maximum, greater: true},
		{name: "maximum to minimum", initial: maximum, new: minimum, less: true},
		{name: "zero to maximum", initial: 0, new: maximum, greater: true},
		{name: "maximum to zero", initial: maximum, new: 0, less: true},
	}

	if minimum < 0 {
		cases = append(cases,
			intConditionalSwapCase[T]{name: "negative to zero", initial: minimum, new: 0, greater: true},
			intConditionalSwapCase[T]{name: "zero to negative", initial: 0, new: minimum, less: true},
			intConditionalSwapCase[T]{name: "negative to negative", initial: minimum, new: minimum + 1, greater: true},
		)
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var value atom.Int[T]

			value.Store(testCase.initial)

			old, swapped := value.SwapIfLess(testCase.new)
			if old != testCase.initial || swapped != testCase.less {
				t.Fatalf("SwapIfLess(%v) = (%v, %v), want (%v, %v)", testCase.new, old, swapped, testCase.initial, testCase.less)
			}

			want := testCase.initial

			if testCase.less {
				want = testCase.new
			}

			got := value.Load()
			if got != want {
				t.Fatalf("Load after SwapIfLess = %v, want %v", got, want)
			}

			value.Store(testCase.initial)

			old, swapped = value.SwapIfGreater(testCase.new)
			if old != testCase.initial || swapped != testCase.greater {
				t.Fatalf("SwapIfGreater(%v) = (%v, %v), want (%v, %v)", testCase.new, old, swapped, testCase.initial, testCase.greater)
			}

			want = testCase.initial

			if testCase.greater {
				want = testCase.new
			}

			got = value.Load()
			if got != want {
				t.Fatalf("Load after SwapIfGreater = %v, want %v", got, want)
			}
		})
	}

	t.Run("overflow", func(t *testing.T) {
		var value atom.Int[T]

		// Add can leave upper bits that are not part of the value of a narrow T.
		value.Store(maximum)
		value.Add(1)

		old, swapped := value.SwapIfLess(minimum)
		if old != minimum || swapped {
			t.Fatalf("SwapIfLess at wrapped minimum = (%v, %v), want (%v, false)", old, swapped, minimum)
		}

		old, swapped = value.SwapIfGreater(minimum + 1)
		if old != minimum || !swapped {
			t.Fatalf("SwapIfGreater after overflow = (%v, %v), want (%v, true)", old, swapped, minimum)
		}

		got := value.Load()
		if got != minimum+1 {
			t.Fatalf("Load after overflow swap = %v, want %v", got, minimum+1)
		}

		value.Store(minimum)
		value.Add(^T(0))

		old, swapped = value.SwapIfGreater(maximum)
		if old != maximum || swapped {
			t.Fatalf("SwapIfGreater at wrapped maximum = (%v, %v), want (%v, false)", old, swapped, maximum)
		}

		old, swapped = value.SwapIfLess(maximum - 1)
		if old != maximum || !swapped {
			t.Fatalf("SwapIfLess after underflow = (%v, %v), want (%v, true)", old, swapped, maximum)
		}

		got = value.Load()
		if got != maximum-1 {
			t.Fatalf("Load after underflow swap = %v, want %v", got, maximum-1)
		}
	})

	t.Run("randomized", testIntConditionalSwapsRandomized[T])
}

func testIntConditionalSwapsRandomized[T testInteger](t *testing.T) {
	var value atom.Int[T]

	random := rand.New(rand.NewPCG(1, 2))
	directions := [...]bool{false, true}

	for range 1000 {
		stored := T(random.Uint64())
		delta := T(random.Uint64())
		candidate := T(random.Uint64())
		initial := stored + delta

		for _, greater := range directions {
			var (
				old     T
				swapped bool
			)

			// Add intentionally permits noncanonical upper bits for narrow types.
			value.Store(stored)
			value.Add(delta)

			shouldSwap := candidate < initial

			if greater {
				old, swapped = value.SwapIfGreater(candidate)
				shouldSwap = candidate > initial
			} else {
				old, swapped = value.SwapIfLess(candidate)
			}

			if old != initial || swapped != shouldSwap {
				t.Fatalf("greater=%v, initial=%v, new=%v: got (%v, %v), want (%v, %v)",
					greater, initial, candidate, old, swapped, initial, shouldSwap)
			}

			want := initial

			if shouldSwap {
				want = candidate
			}

			got := value.Load()
			if got != want {
				t.Fatalf("greater=%v, initial=%v, new=%v: Load = %v, want %v", greater, initial, candidate, got, want)
			}
		}
	}
}

func testIntConditionalSwapsConcurrent(t *testing.T, greater bool) {
	var (
		value     atom.Int[int64]
		waitGroup sync.WaitGroup
		deltas    [intTestGoroutines]int64
	)

	initial := int64(intTestGoroutines*intTestOps + 1)
	want := int64(1)

	if greater {
		initial = 0
		want = intTestGoroutines * intTestOps
	}

	value.Store(initial)

	start := make(chan struct{})
	waitGroup.Add(intTestGoroutines)

	for worker := range intTestGoroutines {
		go func() {
			defer waitGroup.Done()

			<-start

			for operation := range intTestOps {
				var (
					old     int64
					swapped bool
				)

				candidate := int64(operation*intTestGoroutines + worker + 1)

				if greater {
					old, swapped = value.SwapIfGreater(candidate)
				} else {
					candidate = initial - candidate
					old, swapped = value.SwapIfLess(candidate)
				}

				shouldSwap := candidate < old

				if greater {
					shouldSwap = candidate > old
				}

				if swapped != shouldSwap {
					t.Errorf("conditional swap of %v = (%v, %v), want swapped = %v", candidate, old, swapped, shouldSwap)
				}

				if swapped {
					deltas[worker] += candidate - old
				}
			}
		}()
	}

	close(start)
	waitGroup.Wait()

	got := value.Load()
	if got != want {
		t.Fatalf("Load after concurrent swaps = %v, want %v", got, want)
	}

	// Successful swaps must form a chain: their returned old values telescope.
	total := initial

	for _, delta := range deltas {
		total += delta
	}

	if total != got {
		t.Fatalf("sum of successful swap changes = %v, want final value %v", total, got)
	}
}
