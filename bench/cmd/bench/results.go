//go:build go1.24

package main

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type sample struct {
	nanoseconds float64
	bytes       float64
	allocations float64
}

type benchmarkKey struct {
	name string
	cpu  int
}

type pair struct {
	atom   []sample
	stdlib []sample
}

type report struct {
	pairs       map[benchmarkKey]*pair
	environment []string
}

type spread struct {
	median float64
	low    float64
	high   float64
}

func parseResults(input io.Reader) (report, error) {
	result := report{pairs: make(map[benchmarkKey]*pair)}
	scanner := bufio.NewScanner(input)
	complete := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "# complete" {
			complete = true
		}

		if strings.HasPrefix(line, "FAIL") || strings.HasPrefix(line, "--- FAIL") {
			return result, fmt.Errorf("failed benchmark run: %s", line)
		}

		if strings.HasPrefix(line, "# go version ") || strings.HasPrefix(line, "cpu: ") {
			text := strings.TrimPrefix(line, "# ")
			if !slices.Contains(result.environment, text) {
				result.environment = append(result.environment, text)
			}
		}

		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}

		key, implementation, measurement, err := parseSample(line)
		if err != nil {
			return result, err
		}

		values := result.pairs[key]
		if values == nil {
			values = &pair{}
			result.pairs[key] = values
		}

		if implementation == "Atom" {
			values.atom = append(values.atom, measurement)
		} else {
			values.stdlib = append(values.stdlib, measurement)
		}
	}

	err := scanner.Err()
	if err != nil {
		return result, err
	}

	if !complete || len(result.pairs) == 0 {
		return result, errors.New("incomplete or empty benchmark run; collect results with Go 1.24 or later")
	}

	count := 0

	for key, values := range result.pairs {
		if count == 0 {
			count = len(values.atom)
		}

		if len(values.atom) == 0 || len(values.atom) != len(values.stdlib) || len(values.atom) != count {
			return result, fmt.Errorf("unpaired or inconsistent sample counts for %s at CPU=%d", key.name, key.cpu)
		}
	}

	return result, nil
}

func parseSample(line string) (benchmarkKey, string, sample, error) {
	var (
		fields      [16]string
		measurement sample
		key         benchmarkKey
	)

	count := 0

	for field := range strings.FieldsSeq(line) {
		if count == len(fields) {
			return key, "", measurement, fmt.Errorf("too many benchmark fields: %s", line)
		}

		fields[count] = field
		count++
	}

	if count < 8 || count%2 != 0 {
		return key, "", measurement, fmt.Errorf("missing benchmark metrics: %s", line)
	}

	iterations, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil || iterations == 0 {
		return key, "", measurement, fmt.Errorf("invalid iteration count: %s", line)
	}

	name := fields[0]
	key.cpu = 1
	suffix := strings.LastIndexByte(name, '-')

	if suffix >= 0 {
		key.cpu, err = strconv.Atoi(name[suffix+1:])
		if err != nil || key.cpu < 1 {
			return key, "", measurement, fmt.Errorf("invalid CPU suffix: %s", line)
		}

		name = name[:suffix]
	}

	key.name, name, _ = strings.Cut(strings.TrimPrefix(name, "Benchmark"), "/")
	if name != "Atom" && name != "Stdlib" {
		return key, "", measurement, fmt.Errorf("unknown implementation: %s", line)
	}

	seen := uint8(0)

	for index := 2; index < count; index += 2 {
		number, parseErr := strconv.ParseFloat(fields[index], 64)
		if parseErr != nil || number < 0 || math.IsNaN(number) || math.IsInf(number, 0) {
			return key, "", measurement, fmt.Errorf("invalid metric: %s", line)
		}

		switch fields[index+1] {
		case "ns/op":
			measurement.nanoseconds = number
			seen |= 1
		case "B/op":
			measurement.bytes = number
			seen |= 2
		case "allocs/op":
			measurement.allocations = number
			seen |= 4
		}
	}

	if seen != 7 || measurement.nanoseconds <= 0 {
		return key, "", measurement, fmt.Errorf("missing timing or allocation metrics: %s", line)
	}

	return key, name, measurement, nil
}

func writeResults(directory string, result report) error {
	keys := make([]benchmarkKey, 0, len(result.pairs))

	for key := range result.pairs {
		keys = append(keys, key)
	}

	slices.SortFunc(keys, func(left, right benchmarkKey) int {
		order := strings.Compare(left.name, right.name)
		if order != 0 {
			return order
		}

		return cmp.Compare(left.cpu, right.cpu)
	})

	tables := make(map[string]*bytes.Buffer, 8)

	var summary bytes.Buffer

	summary.Grow(256 * len(keys))
	fmt.Fprintln(&summary, "benchmark\tcpu\tsamples\tatom_ns\tstdlib_ns\tatom/stdlib\tatom_B\tstdlib_B\tatom_allocs\tstdlib_allocs")

	for _, key := range keys {
		values := result.pairs[key]
		atomTime, atomBytes, atomAllocs := summarize(values.atom)
		stdlibTime, stdlibBytes, stdlibAllocs := summarize(values.stdlib)
		ratio := atomTime.median / stdlibTime.median
		fmt.Fprintf(&summary, "%s\t%d\t%d\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\n",
			key.name, key.cpu, len(values.atom), atomTime.median, stdlibTime.median, ratio,
			atomBytes.median, stdlibBytes.median, atomAllocs.median, stdlibAllocs.median)

		filename, label := tableName(key)

		table := tables[filename]
		if table == nil {
			table = &bytes.Buffer{}
			table.Grow(256 * len(keys))
			fmt.Fprintln(table, "# label atom_ns atom_min atom_max stdlib_ns stdlib_min stdlib_max atom_B stdlib_B atom_allocs stdlib_allocs ratio")
			tables[filename] = table
		}

		fmt.Fprintf(table, "%s\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\t%.6g\n",
			label, atomTime.median, atomTime.low, atomTime.high, stdlibTime.median, stdlibTime.low, stdlibTime.high,
			atomBytes.median, stdlibBytes.median, atomAllocs.median, stdlibAllocs.median, ratio)
	}

	expected := []string{"integer", "conditional", "value", "parallel_Load", "parallel_Add", "parallel_CASIncrement", "parallel_ConditionalPair"}

	for _, name := range expected {
		if tables[name] == nil {
			return fmt.Errorf("missing benchmark group %s", name)
		}
	}

	for name, table := range tables {
		err := os.WriteFile(filepath.Join(directory, name+".tsv"), table.Bytes(), 0o644)
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(filepath.Join(directory, "summary.tsv"), summary.Bytes(), 0o644)
	if err != nil {
		return err
	}

	var environment bytes.Buffer

	fmt.Fprintf(&environment, "benchmark_environment = %s\n", strconv.Quote(strings.Join(result.environment, " | ")))

	return os.WriteFile(filepath.Join(directory, "environment.gnuplot"), environment.Bytes(), 0o644)
}

func tableName(key benchmarkKey) (string, string) {
	name, matched := strings.CutPrefix(key.name, "Parallel")
	if matched {
		return "parallel_" + name, strconv.Itoa(key.cpu)
	}

	name, matched = strings.CutPrefix(key.name, "Conditional")
	if matched {
		return "conditional", strconv.Quote(name)
	}

	name, matched = strings.CutPrefix(key.name, "Value")
	if matched {
		return "value", strconv.Quote(name)
	}

	return "integer", strconv.Quote(key.name)
}

func summarize(values []sample) (spread, spread, spread) {
	nanoseconds := make([]float64, len(values))
	bytes := make([]float64, len(values))
	allocations := make([]float64, len(values))

	for index, value := range values {
		nanoseconds[index] = value.nanoseconds
		bytes[index] = value.bytes
		allocations[index] = value.allocations
	}

	return distribution(nanoseconds), distribution(bytes), distribution(allocations)
}

func distribution(values []float64) spread {
	slices.Sort(values)
	middle := len(values) / 2
	median := values[middle]

	if len(values)%2 == 0 {
		median = (values[middle-1] + median) / 2
	}

	return spread{median: median, low: values[0], high: values[len(values)-1]}
}
