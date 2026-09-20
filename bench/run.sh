#!/usr/bin/env bash
# Run from the repository root. Requires Go 1.24+, Bash, awk, sort and optionally gnuplot.
set -euo pipefail
export LC_ALL=C

benchtime=${BENCHTIME:-500ms}
cpus=${CPU:-1,2,4,8}
plot=${PLOT:-true}
results=bench/results

if [[ "$plot" == true ]]; then
    command -v gnuplot >/dev/null
fi

mkdir -p "$results"
trap 'rm -f "$results/raw.tmp" "$results/tables.tmp"' EXIT

{
    printf '# Recorded: %s\n' "$(date -u +%FT%TZ)"
    go version
    go env -json GOOS GOARCH GOAMD64 GOARM GOARM64 CGO_ENABLED GOEXPERIMENT GOFLAGS
    printf '# benchtime=%s cpu=%s GODEBUG=%s GOMAXPROCS=%s\n' "$benchtime" "$cpus" "${GODEBUG:-}" "${GOMAXPROCS:-}"
    go test ./bench -run='^$' -bench='^Benchmark(Uint64|Int32|Conditional|Value)' -benchmem -benchtime="$benchtime" -count=1 -cpu=1
    go test ./bench -run='^$' -bench='^BenchmarkParallel' -benchmem -benchtime="$benchtime" -count=1 -cpu="$cpus"
} | tee "$results/raw.tmp"

awk -f bench/results.awk "$results/raw.tmp" | sort -k1,1 -k2,2n -k2,2 > "$results/tables.tmp"
awk -v directory="$results" '
    {
        file = directory "/" $1 ".tsv"
        if (!written[file]++) print "# case\tatom_ns\tstdlib_ns\tatom_B\tstdlib_B\tatom_allocs\tstdlib_allocs" > file
        sub(/^[^\t]*\t/, "")
        print >> file
    }
' "$results/tables.tmp"
mv "$results/raw.tmp" "$results/raw.txt"

if [[ "$plot" == true ]]; then
    (cd "$results" && gnuplot ../benchmark.gnuplot)
fi
