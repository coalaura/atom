//go:build go1.24

package main

import (
	"strings"
	"testing"
)

func TestParseResults(t *testing.T) {
	input := "# go version go1.27.1 linux/amd64\n" +
		"cpu: example CPU\n" +
		"BenchmarkParallelConditionalPair/Atom-8 1000 12.5 ns/op 2 calls/op 0 B/op 0 allocs/op\n" +
		"BenchmarkParallelConditionalPair/Stdlib-8 1000 10 ns/op 2 calls/op 0 B/op 0 allocs/op\n" +
		"BenchmarkUint64Load/Atom 1000 2 ns/op 0 B/op 0 allocs/op\n" +
		"BenchmarkUint64Load/Stdlib 1000 1 ns/op 0 B/op 0 allocs/op\n" +
		"# complete\n"

	result, err := parseResults(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	parallel := result.pairs[benchmarkKey{name: "ParallelConditionalPair", cpu: 8}]
	if parallel == nil || parallel.atom[0].nanoseconds != 12.5 || parallel.stdlib[0].nanoseconds != 10 {
		t.Fatalf("incorrect parallel pair: %+v", parallel)
	}

	serial := result.pairs[benchmarkKey{name: "Uint64Load", cpu: 1}]
	if serial == nil || serial.atom[0].nanoseconds != 2 || serial.stdlib[0].nanoseconds != 1 {
		t.Fatalf("incorrect serial pair: %+v", serial)
	}

	if len(result.environment) != 2 {
		t.Fatalf("environment = %v", result.environment)
	}
}

func TestParseResultsRejectsInvalidRuns(t *testing.T) {
	valid := "BenchmarkUint64Load/Atom 1000 2 ns/op 0 B/op 0 allocs/op\n" +
		"BenchmarkUint64Load/Stdlib 1000 1 ns/op 0 B/op 0 allocs/op\n"
	cases := map[string]string{
		"empty":           "# complete\n",
		"unfinished":      valid,
		"failed":          valid + "FAIL\n# complete\n",
		"missing pair":    "BenchmarkUint64Load/Atom 1000 2 ns/op 0 B/op 0 allocs/op\n# complete\n",
		"missing metrics": strings.ReplaceAll(valid, "0 allocs/op", "") + "# complete\n",
		"zero time":       strings.ReplaceAll(valid, "2 ns/op", "0 ns/op") + "# complete\n",
		"nonfinite time":  strings.ReplaceAll(valid, "2 ns/op", "NaN ns/op") + "# complete\n",
		"different CPU":   strings.ReplaceAll(valid, "Stdlib ", "Stdlib-8 ") + "# complete\n",
		"different count": valid + "BenchmarkUint64Load/Atom 1000 2 ns/op 0 B/op 0 allocs/op\n# complete\n",
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := parseResults(strings.NewReader(input))
			if err == nil {
				t.Fatal("accepted an incomplete or invalid comparison")
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	values := []sample{
		{nanoseconds: 9, bytes: 32, allocations: 2},
		{nanoseconds: 1, bytes: 0, allocations: 0},
		{nanoseconds: 3, bytes: 16, allocations: 1},
		{nanoseconds: 5, bytes: 16, allocations: 1},
	}
	nanoseconds, bytes, allocations := summarize(values)

	if nanoseconds != (spread{median: 4, low: 1, high: 9}) {
		t.Fatalf("timing = %+v", nanoseconds)
	}

	if bytes != (spread{median: 16, low: 0, high: 32}) || allocations != (spread{median: 1, low: 0, high: 2}) {
		t.Fatalf("allocations = %+v bytes, %+v allocs", bytes, allocations)
	}

	odd, _, _ := summarize(values[:3])
	if odd.median != 3 || values[0].nanoseconds != 9 {
		t.Fatalf("odd median = %g; original samples must remain unchanged", odd.median)
	}
}
