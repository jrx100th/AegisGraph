# Benchmark commands

Run the deterministic microbenchmarks with:

    go test -run '^$' -bench '^Benchmark' -benchmem ./...

The benchmarks measure the current synthetic analysis and network prerequisite functions. They do not claim 50,000-node graph performance. Larger graph benchmarks remain required before making scale claims.
