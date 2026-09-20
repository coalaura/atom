# Run from bench/results: gnuplot ../benchmark.gnuplot
# Tables contain one paired sample per case; raw.txt records the environment.
reset
set encoding utf8
set terminal pngcairo size 1600,950 font "DejaVu Sans,11" noenhanced

set border 3
set tics nomirror
set grid ytics back lc rgb "#dddddd"
set key outside top center horizontal
set style fill solid 0.85 border -1
set style histogram clustered gap 2
set style data histograms
set boxwidth 0.85
set yrange [0:*]
set xtics rotate by 35 right
set bmargin 10
set ylabel "ns/op (lower is better)"
set label 1 "Environment and measurement settings: raw.txt" at screen 0.02,0.025 left font ",9"

do for [family in "integer conditional value"] {
    set output sprintf("benchmark_%s.png", family)
    heading = family eq "integer" ? "Integer operations" : family eq "conditional" ? "Conditional swaps (stdlib: Load/CAS loop)" : "Value operations (typed results on both sides)"
    set title heading
    plot sprintf("%s.tsv", family) using 2:xticlabels(1) title "coalaura/atom" lc rgb "#3377bb", \
         "" using 3 title "sync/atomic" lc rgb "#ee8833"
}

set output "benchmark_parallel.png"
unset label 1
set bmargin
set xtics 1 norotate
set xlabel "GOMAXPROCS (workers)"
set ylabel "aggregate ns/op (lower is better)"
set style data linespoints
set pointsize 1
set multiplot layout 2,2 title "Shared-state throughput"

do for [operation in "Load Add CASIncrement ConditionalPair"] {
    set title operation eq "ConditionalPair" ? "ConditionalPair (two calls per operation)" : operation
    plot sprintf("parallel_%s.tsv", operation) using 1:2 title "coalaura/atom" lw 2 pt 7 lc rgb "#3377bb", \
         "" using 1:3 title "sync/atomic" lw 2 pt 5 lc rgb "#ee8833"
}
unset multiplot

set output "benchmark_allocations.png"
set style data histograms
set style histogram clustered gap 2
set xtics autofreq rotate by 35 right
unset xlabel
set bmargin 8
set multiplot layout 2,1 title "Value allocations (lower is better)"
unset title
set ylabel "B/op"
plot "value.tsv" using 4:xticlabels(1) title "coalaura/atom" lc rgb "#3377bb", \
     "" using 5 title "sync/atomic" lc rgb "#ee8833"
set ylabel "allocs/op"
plot "value.tsv" using 6:xticlabels(1) title "coalaura/atom" lc rgb "#3377bb", \
     "" using 7 title "sync/atomic" lc rgb "#ee8833"
unset multiplot
unset output
