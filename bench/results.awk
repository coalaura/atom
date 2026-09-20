# Pair the single samples emitted by run.sh. The measured Go loops remain direct calls.
/^Benchmark/ {
    split($1, parts, "/")
    name = substr(parts[1], 10)
    implementation = parts[2]
    cpu = 1
    if (match(implementation, /-[0-9]+$/)) {
        cpu = substr(implementation, RSTART + 1)
        implementation = substr(implementation, 1, RSTART - 1)
    }
    if (implementation != "Atom" && implementation != "Stdlib") exit 1

    key = name SUBSEP cpu
    sample = key SUBSEP implementation
    if (seen[sample]++) exit 1
    nanos[sample] = bytes[sample] = allocations[sample] = -1
    for (field = 3; field < NF; field += 2) {
        if ($(field + 1) == "ns/op") nanos[sample] = $field
        if ($(field + 1) == "B/op") bytes[sample] = $field
        if ($(field + 1) == "allocs/op") allocations[sample] = $field
    }
    if (nanos[sample] <= 0 || bytes[sample] < 0 || allocations[sample] < 0) exit 1
    cases[key] = name
    cpus[key] = cpu
    count++
}
END {
    if (!count) exit 1
    OFS = "\t"
    for (key in cases) {
        atom = key SUBSEP "Atom"
        stdlib = key SUBSEP "Stdlib"
        if (!seen[atom] || !seen[stdlib]) exit 1
        name = cases[key]
        family = "integer"
        label = name
        if (sub(/^Conditional/, "", label)) family = "conditional"
        else if (sub(/^Value/, "", label)) family = "value"
        else if (sub(/^Parallel/, "", label)) {
            family = "parallel_" label
            label = cpus[key]
        }
        print family, label, nanos[atom], nanos[stdlib], bytes[atom], bytes[stdlib], allocations[atom], allocations[stdlib]
    }
}
