# atom

Generic atomic values and integers.

```go
var v atom.Value[int]
v.Store(42)
n := v.Load() // int, not any

var counter atom.Int[int]
counter.Add(1)
```

Drop-in semantics of `sync/atomic.Value`, typed with generics so callers don't cast.

## Install

```bash
go get github.com/coalaura/atom
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/coalaura/atom"
)

func main() {
	var v atom.Value[string]
	v.Store("hello")
	fmt.Println(v.Load())

	prev := v.Swap("world")
	fmt.Println(prev, v.Load())

	ok := v.CompareAndSwap("world", "!")
	fmt.Println(ok, v.Load())
}
```

Same rules as `sync/atomic.Value`:

- Do not copy a `Value` after first use
- All stored values must share one concrete type
- Store/Swap/CAS of a nil interface panics

### Conditional integer swaps

`Int[T].SwapIfLess(new)` replaces the current value only when `new` is less; `SwapIfGreater(new)` replaces it only when `new` is greater. Both atomically return the previous value and whether the swap occurred. Equal values are not swapped and comparisons respect the signedness and width of `T`.

```go
var limit atom.Int[int]
limit.Store(10)

old, swapped := limit.SwapIfLess(5)   // 10, true; limit is now 5
old, swapped = limit.SwapIfGreater(3) // 5, false; limit remains 5
```

The zero value starts at zero, so initialize it appropriately when tracking a minimum or maximum. Conditional swaps perform their comparisons and any required retries inside architecture-specific assembly. Race builds use instrumented `sync/atomic` operations.

## License

[BSD](https://github.com/golang/go/blob/master/LICENSE)
