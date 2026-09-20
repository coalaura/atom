# Upstream uint64 atomics

The architecture-specific `atomic_*.go` and `atomic_*.s` files, `atomic_andor_generic.go`, `sys_*_arm.s`, `stubs.go`, and `unaligned.go` are trimmed copies of `src/internal/runtime/atomic` from the locally installed Go toolchain (`go1.27.1`, `C:\Program Files\Go\src`). Retained declarations, implementations, comments, compiler directives, and assembly instructions follow that source verbatim. The race adapter and tests are local code.

The retained operations are `Load64`, `Store64`, `Xadd64`, `Xchg64`, `Cas64`, `And64`, and `Or64`, plus their implementation dependencies. In particular, ARM needs 32-bit CAS/store, its native 64-bit helpers, and its original striped spinlocks; MIPS needs its original spinlock. These helpers are necessary for the uint64 operations even though some operate on 32-bit lock words.

## Integration changes

- The package name is `internal`, and Go imports of `internal/cpu` point to this module's `internal/cpu` package.
- Build constraints add `!race`; `stubs.go` is restricted to ARM, which is the only retained implementation needing its 32-bit CAS declaration.
- Unrelated operations and their declarations, imports, and comments are omitted. File names and the upstream architecture and operating-system selection are retained.
- `api_decl.go` and `asm.s` provide the `sync/atomic`-style uint64 names used by `Int`. `atomic_race.go` delegates those names to `sync/atomic` so race builds remain instrumented.

The `cpu` subdirectory contains the upstream cache-line padding constants and feature-structure layouts needed by `unsafe.Offsetof`. These copies only calculate offsets: the unchanged assembly reads the runtime's actual `internal/cpu` variables, including feature detection and `GODEBUG` overrides. Keep the layouts and atomic source synchronized when updating the snapshot. Go's compiler intrinsics apply to the standard-library package path; this copy uses the retained Go and assembly implementations.

## Architecture coverage

The snapshot includes every architecture handled by upstream: `386`, `amd64`, `arm`, `arm64`, `loong64`, `mips`, `mipsle`, `mips64`, `mips64le`, `ppc64`, `ppc64le`, `riscv64`, `s390x`, and `wasm`. It preserves ARM Linux/non-Linux helpers, ARM v5/v6 fallbacks and v7 instructions, ARM64 LSE dispatch, and LoongArch CPU-feature dispatch.

The verified toolchain is Go 1.27.1. Older-toolchain compatibility cannot be inferred from the module's language version: the assembler must support the copied instructions, and the runtime CPU layouts must match the copied offsets, especially on LoongArch.

## Maintenance

Preserve upstream source formatting and comments when refreshing these files; do not restyle copied source to satisfy house `vet` rules. Tests follow the repository's `vet` style. Verify retained functions and assembly blocks against the source snapshot, run the tests, and compile and link test binaries for every architecture, including both ARM OS branches and both WebAssembly targets.
