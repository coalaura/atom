# Upstream uint64 atomics

The original operations in the architecture-specific `atomic_*.go` and `atomic_*.s` files, `atomic_andor_generic.go`, `sys_*_arm.s`, `stubs.go` and `unaligned.go` are trimmed copies of `src/internal/runtime/atomic` from the locally installed Go toolchain (`go1.27.1`, `C:\Program Files\Go\src`). Retained declarations, implementations, comments, compiler directives and assembly instructions follow that source verbatim. Conditional swaps, the race adapter and tests are local code.

The retained operations are `Load64`, `Store64`, `Xadd64`, `Xchg64`, `Cas64`, `And64` and `Or64`, plus their implementation dependencies. In particular, ARM needs 32-bit CAS/store, its native 64-bit helpers and its original striped spinlocks; MIPS needs its original spinlock. These helpers are necessary for the uint64 operations even though some operate on 32-bit lock words.

## Integration changes

- The package name is `internal` and Go imports of `internal/cpu` point to this module's `internal/cpu` package.
- Build constraints add `!race`; `stubs.go` is restricted to ARM, which is the only retained implementation needing its 32-bit CAS declaration.
- Unrelated operations and their declarations, imports and comments are omitted. File names and the upstream architecture and operating-system selection are retained.
- `api_decl.go` and `asm.s` provide the `sync/atomic`-style uint64 names used by `Int`. `atomic_race.go` delegates those names to `sync/atomic` so race builds remain instrumented.

The `cpu` subdirectory contains the upstream cache-line padding constants and feature-structure layouts needed by `unsafe.Offsetof`. These copies only calculate offsets: the unchanged assembly reads the runtime's actual `internal/cpu` variables, including feature detection and `GODEBUG` overrides. Keep the layouts and atomic source synchronized when updating the snapshot. Go's compiler intrinsics apply to the standard-library package path; this copy uses the retained Go and assembly implementations.

## Architecture coverage

The snapshot includes every architecture handled by upstream: `386`, `amd64`, `arm`, `arm64`, `loong64`, `mips`, `mipsle`, `mips64`, `mips64le`, `ppc64`, `ppc64le`, `riscv64`, `s390x` and `wasm`. It preserves ARM Linux/non-Linux helpers, ARM v5/v6 fallbacks and v7 instructions, ARM64 LSE dispatch and LoongArch CPU-feature dispatch.

The verified toolchain is Go 1.27.1. Older-toolchain compatibility cannot be inferred from the module's language version: the assembler must support the copied instructions and the runtime CPU layouts must match the copied offsets, especially on LoongArch.

## Conditional swaps

`SwapIfLessUint64` and `SwapIfGreaterUint64` are local additions integrated into the corresponding `atomic_*.s` files. `api_decl.go` declares their common ABI: an address, the full new value, a width mask and a sign bit. Comparing unsigned keys `(value ^ sign) & mask` reproduces the ordering of the caller's integer type, including negative values and narrow values whose backing storage contains overflow bits. Greater-than reverses this ordering by complementing the sign argument within the mask and tail-calling the same assembly loop. Equal keys never cause a store; successful swaps return and compare the full original bits.

The implementations use native CAS or load-exclusive/store-exclusive loops, preserving atomic memory ordering on both successful and read-only outcomes. ARM64 and LoongArch use their baseline exclusive instructions without requiring optional CPU extensions. Pre-v7 ARM and 32-bit MIPS share the upstream operations' locks; comparisons and updates stay in assembly. ARM's `lockSwap64` helper acquires the original striped lock. WebAssembly uses an uninterrupted load/compare/store sequence, matching upstream's single-threaded execution assumption.

The added assembly follows the surrounding source's function-signature comments, `Atomically:` pseudocode, instruction formatting and architecture-specific calling conventions. There is no upstream implementation of these conditional operations to copy verbatim; their generic width/sign ABI and implementations are specific to this package.

## Race builds and testing

Go's standard library also selects different implementations under `-race`. In the source snapshot, `src/sync/atomic/asm.s` is guarded by `!race`. `src/runtime/race_*.s` supplies race-aware `sync/atomic` entry points; for example, the amd64 `SwapUint64` entry point reaches `__tsan_go_atomic64_exchange` through `SwapInt64`. The compiler's `findIntrinsic` in `src/cmd/compile/internal/ssagen/intrinsics.go` disables `sync/atomic` intrinsics under `-race` so those calls reach the race runtime. This special handling does not apply to arbitrary package paths or copied assembly.

`atomic_race.go` therefore delegates the ordinary operations to `sync/atomic` and expresses conditional swaps using its instrumented loads and compare-and-swaps. This lets the detector observe both atomic accesses and the synchronization they establish. Removing the adapter would leave native assembly accesses invisible to the detector; merely running that assembly in a race build would not validate its atomicity. Acquire/release annotations alone do not provide equivalent atomic-access instrumentation.

Run both `go test -count=1 ./...` and `go test -race -count=1 ./...`. The first exercises the actual native implementation, including concurrent updates, returned-value checks and signed/narrow-value semantics. The second exercises the race-aware implementation and detects races in instrumented code using it. Neither replaces the other. Cross-compiling and linking verifies architecture selection and assembly linkage, but executing the tests on each target is still needed to validate its behavior.

## Maintenance

Preserve upstream source formatting and comments when refreshing these files; do not restyle copied source to satisfy house `vet` rules. Tests follow the repository's `vet` style. Verify retained functions and assembly blocks against the source snapshot, run the tests and compile and link test binaries for every architecture, including both ARM OS branches and both WebAssembly targets.
