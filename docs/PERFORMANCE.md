# Performance

GitHub Actions run 34030372218, job 101478659148, passed the benchmark command `go test -run "^$" -bench "^Benchmark" -benchmem ./...` on the hosted Go 1.22 runner. Measured output:

- BenchmarkDemoSnapshot: 3,208 ns/op, 5,936 B/op, 39 allocs/op.
- BenchmarkNetworkReachability: 100.5 ns/op, 112 B/op, 3 allocs/op.
- BenchmarkReplayCollection: 5,193 ns/op, 7,848 B/op, 63 allocs/op.

These are small deterministic workloads, not large-graph or production-latency claims. The replay benchmark measures the simulated collector, normalization, and analysis path for the internet-sensitive S3 scenario.

The larger benchmark plan is to generate deterministic graphs of 10,000 nodes/50,000 edges and 50,000 nodes/250,000 edges, then measure persistence, loading, filtered graph queries, rule evaluation, path traversal, blast-radius traversal, API latency, memory, and SQLite size. Those measurements have not yet been executed.

A benchmark record must include command, hardware, Go version, dataset seed, and raw output. Distributed infrastructure is not an acceptable substitute for fixing accidental quadratic behavior.
