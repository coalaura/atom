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

## License

[BSD](https://github.com/golang/go/blob/master/LICENSE)
