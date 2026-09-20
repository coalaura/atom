//go:build go1.24

// Command bench records paired benchmarks and renders their results with gnuplot.
// Run it from the repository root: go run ./bench/cmd/bench.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type options struct {
	output     string
	benchtime  string
	cpus       string
	gnuplot    string
	count      int
	plot       bool
	renderOnly bool
}

type benchmarkRun struct {
	pattern string
	cpus    string
}

func main() {
	var config options

	flag.StringVar(&config.output, "out", "bench/results", "output directory")
	flag.StringVar(&config.benchtime, "benchtime", "200ms", "time or iteration count per benchmark sample")
	flag.StringVar(&config.cpus, "cpu", "1,2,4,8", "GOMAXPROCS values for parallel benchmarks")
	flag.StringVar(&config.gnuplot, "gnuplot", "gnuplot", "gnuplot executable")
	flag.IntVar(&config.count, "count", 5, "samples per implementation and CPU count")
	flag.BoolVar(&config.plot, "plot", true, "render PNGs after writing the data tables")
	flag.BoolVar(&config.renderOnly, "render-only", false, "rebuild tables and plots from the saved raw.txt")
	flag.Parse()

	err := run(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(config options) error {
	if config.count < 1 {
		return errors.New("count must be positive")
	}

	_, err := os.Stat("bench/benchmark.gnuplot")
	if err != nil {
		return fmt.Errorf("run from the atom repository root: %w", err)
	}

	if config.plot {
		config.gnuplot, err = exec.LookPath(config.gnuplot)
		if err != nil {
			return fmt.Errorf("gnuplot is required for rendering; use -plot=false to collect data only: %w", err)
		}

		config.gnuplot, err = filepath.Abs(config.gnuplot)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(config.output, 0o755)
	if err != nil {
		return err
	}

	if !config.renderOnly {
		err = record(config)
		if err != nil {
			return err
		}
	}

	raw, err := os.Open(filepath.Join(config.output, "raw.txt"))
	if err != nil {
		return err
	}

	results, parseErr := parseResults(raw)

	err = errors.Join(parseErr, raw.Close())
	if err != nil {
		return err
	}

	err = writeResults(config.output, results)
	if err != nil {
		return err
	}

	if config.plot {
		script, pathErr := filepath.Abs("bench/benchmark.gnuplot")
		if pathErr != nil {
			return pathErr
		}

		command := exec.Command(config.gnuplot, script)
		command.Dir = config.output
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		err = command.Run()
		if err != nil {
			return fmt.Errorf("render plots: %w", err)
		}
	}

	fmt.Fprintln(os.Stdout, "Benchmark results:", config.output)

	return nil
}

func record(config options) error {
	// Reject duplicate CPU counts: otherwise they silently double the sample
	// count for only some pairs and make the summary misleading.
	seen := make(map[int]bool)

	for text := range strings.SplitSeq(config.cpus, ",") {
		count, err := strconv.Atoi(text)
		if err != nil || count < 1 || seen[count] {
			return fmt.Errorf("invalid or duplicate CPU count %q", text)
		}

		seen[count] = true
	}

	output, err := os.Create(filepath.Join(config.output, "raw.txt"))
	if err != nil {
		return err
	}

	err = recordCommands(config, io.MultiWriter(os.Stdout, output))
	return errors.Join(err, output.Close())
}

func recordCommands(config options, output io.Writer) error {
	_, err := fmt.Fprintln(output, "# recorded:", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}

	version, err := exec.Command("go", "version").Output()
	if err != nil {
		return err
	}

	fmt.Fprintf(output, "# %s", version)

	environment, err := exec.Command("go", "env", "-json", "GOOS", "GOARCH", "GOAMD64", "GOARM", "GOARM64", "GOEXPERIMENT", "CGO_ENABLED", "GOFLAGS").Output()
	if err != nil {
		return err
	}

	for line := range strings.SplitSeq(strings.TrimSpace(string(environment)), "\n") {
		fmt.Fprintln(output, "#", line)
	}

	fmt.Fprintf(output, "# GODEBUG=%q GOMAXPROCS=%q\n", os.Getenv("GODEBUG"), os.Getenv("GOMAXPROCS"))

	runs := []benchmarkRun{
		{pattern: "^Benchmark(Uint64|Int32|Conditional|Value)", cpus: "1"},
		{pattern: "^BenchmarkParallel", cpus: config.cpus},
	}

	// Repeat the whole suite so the two implementations are measured next to
	// each other on every repetition, rather than batching one implementation.
	for repetition := 0; repetition < config.count; repetition++ {
		for _, run := range runs {
			arguments := []string{
				"test", "./bench", "-run=^$", "-bench=" + run.pattern, "-benchmem",
				"-benchtime=" + config.benchtime, "-count=1", "-cpu=" + run.cpus, "-timeout=30m",
			}
			fmt.Fprintf(output, "# sample %d: go %s\n", repetition+1, strings.Join(arguments, " "))

			command := exec.Command("go", arguments...)
			command.Stdout = output
			command.Stderr = output

			err = command.Run()
			if err != nil {
				return fmt.Errorf("benchmark sample %d: %w", repetition+1, err)
			}
		}
	}

	_, err = fmt.Fprintln(output, "# complete")
	return err
}
