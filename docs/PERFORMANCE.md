# Performance

The repository retains a Go replay microbenchmark for regression. Historical Task 2 hosted measurements were:

- BenchmarkDemoSnapshot: 3,208 ns/op, 5,936 B/op, 39 allocs/op.
- BenchmarkNetworkReachability: 100.5 ns/op, 112 B/op, 3 allocs/op.
- BenchmarkReplayCollection: 5,193 ns/op, 7,848 B/op, 63 allocs/op.

Those values describe the previous small workload and are not reused as a Task 3 claim. Run current measurements with:

    go test -run '^$' -bench '^Benchmark' -benchmem ./...

Meaningful scale benchmarks for 10k/50k normalized-resource ingestion, persistence, graph construction, rule evaluation, attack paths, blast radius, repeated rescans, and finding lifecycle remain a verification item. Record command, hardware, Go version, dataset seed, and raw output before reporting numbers.
