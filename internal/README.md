# Conditional integer atomics

This package implements `Int.SwapIfLess` and `Int.SwapIfGreater`. Ordinary integer operations call `sync/atomic` directly from `Int`, allowing compiler intrinsics and using the standard library's architecture support. The previous copies of ordinary atomic operations, their wrappers, private locks and unused CPU layouts have been removed.

The conditional assembly is local code, organized in architecture-specific `atomic_*.s` files using Go assembly conventions. There is no corresponding upstream conditional-swap implementation to copy verbatim.

## Retained Go source

The remaining copied support comes from the locally installed Go toolchain (`go1.27.1`, `C:\Program Files\Go\src`):

- `atomic_arm.s` retains the `CHECK_ALIGN` macro and its stack-unwinding explanation from `src/internal/runtime/atomic/atomic_arm.s`.
- `atomic_arm.go` retains the ARM feature-offset constant from `src/internal/runtime/atomic/atomic_arm.go`, with its CPU import redirected to this module's `internal/cpu` package.
- `cpu/cpu.go` and `cpu/cpu_arm.go` retain the ARM feature layout and cache-line padding from `src/internal/cpu`. They calculate offsets only; assembly reads the runtime's actual `internal/cpu.ARM` variable, including runtime feature detection and `GODEBUG` overrides.
- `unaligned.go` retains the alignment panic from `src/internal/runtime/atomic/unaligned.go`, restricted to the 386 and ARM non-race implementations that use it.

These files keep the Go Authors copyright notices and the retained upstream source formatting. Files containing only local code carry the local copyright notice. Package names and build constraints are adapted for this module; the repository's BSD license covers both attributions.

## Architecture coverage

Conditional swaps cover every architecture handled by Go's runtime atomics: `386`, `amd64`, `arm`, `arm64`, `loong64`, `mips`, `mipsle`, `mips64`, `mips64le`, `ppc64`, `ppc64le`, `riscv64`, `s390x` and `wasm`. ARM retains v5/v6 fallbacks and v7 atomics. The standard library supplies platform-specific support for ordinary operations.

The verified toolchain is Go 1.27.1. Keep the copied ARM CPU layout synchronized with the runtime layout when updating toolchains; other architectures no longer depend on copied runtime CPU layouts.

## Conditional swaps

On architectures other than amd64, `api_decl.go` declares `SwapIfLessUint64` and `SwapIfGreaterUint64` with a common ABI: an address, the full new value, a width mask and a sign bit. Comparing unsigned keys `(value ^ sign) & mask` reproduces the ordering of the caller's integer type, including negative values and narrow values whose backing storage contains overflow bits. Greater-than reverses this ordering by complementing the sign argument within the mask and tail-calling the same assembly loop. Equal keys never cause a store; successful swaps return and compare the full original bits.

On amd64, `Int` instead calls `SwapIfLessInteger` / `SwapIfGreaterInteger`, declared in `conditional_amd64.go`. Their compact ABI passes the type's byte width and signedness and returns only the observed old bits. The Go wrapper derives `swapped` from the strict typed comparison with that old value, which describes both successful and rejected operations without another load. These wrappers inline with the verified toolchain. Assembly uses direct signed/unsigned comparisons for 64-bit types; narrow types shift away overflow bits and move the type's sign bit to bit 63 before comparison. CAS always uses the full original bits. Greater-than has its own loop, avoiding the common ABI's sign transformation and tail jump. Contention retries remain entirely in assembly.

The implementations use native CAS or load-exclusive/store-exclusive loops, preserving atomic memory ordering on both successful and read-only outcomes. ARM64 and LoongArch use their baseline exclusive instructions without requiring optional CPU extensions. ARM without v7 atomics and 32-bit MIPS tail-call `goSwapIfLess64`, which uses `sync/atomic` loads and CAS. These fallbacks must share the standard library's locking protocol to interoperate with ordinary operations. ARM retains the native exclusive loop when v7 atomics are available. WebAssembly uses an uninterrupted load/compare/store sequence, matching upstream's single-threaded execution assumption; its nil-address path calls `sync/atomic.LoadUint64` to raise a recoverable panic.

Assembly uses function-signature comments, `Atomically:` pseudocode and architecture-specific calling conventions. Its comparison metadata and full-width backing representation are specific to this package.

## Race builds and testing

Go's standard library also selects different implementations under `-race`. In the source snapshot, `src/sync/atomic/asm.s` is guarded by `!race`. `src/runtime/race_*.s` supplies race-aware `sync/atomic` entry points; for example, the amd64 `SwapUint64` entry point reaches `__tsan_go_atomic64_exchange` through `SwapInt64`. The compiler's `findIntrinsic` in `src/cmd/compile/internal/ssagen/intrinsics.go` disables `sync/atomic` intrinsics under `-race` so those calls reach the race runtime. This special handling does not apply to arbitrary package paths or copied assembly.

`atomic_race.go` therefore expresses conditional swaps using instrumented `sync/atomic` loads and compare-and-swaps. Ordinary methods already call `sync/atomic` directly. This lets the detector observe both atomic accesses and the synchronization they establish. Removing the adapter would leave native assembly accesses invisible to the detector; merely running that assembly in a race build would not validate its atomicity. Acquire/release annotations alone do not provide equivalent atomic-access instrumentation.

Run both `go test -count=1 ./...` and `go test -race -count=1 ./...`. The first exercises the actual native implementation, including concurrent updates, returned-value checks and signed/narrow-value semantics. The second exercises the race-aware implementation and detects races in instrumented code using it. Neither replaces the other. Cross-compiling and linking verifies architecture selection and assembly linkage, but executing the tests on each target is still needed to validate its behavior.

## Maintenance

Preserve upstream source formatting and comments when refreshing the remaining copied support; do not restyle it to satisfy house `vet` rules. Tests follow the repository's `vet` style. The public-operation tests in `int_atomic_test.go` compare against stdlib atomics and exercise concurrent loads, stores and bitwise operations; `int_swap_test.go` covers typed conditional swaps and mixed-operation concurrency. Run the tests and compile and link test binaries for every architecture, including ARM v5/v6/v7, Linux/non-Linux ARM and both WebAssembly targets.
