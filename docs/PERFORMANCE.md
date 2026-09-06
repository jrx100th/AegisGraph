# Performance

Previously measured on a hosted Go 1.22 CI runner:

- BenchmarkDemoSnapshot: 750.4 ns/op, 3680 B/op, 16 allocs/op.
- BenchmarkNetworkReachability: 63.36 ns/op, 112 B/op, 3 allocs/op.

This task adds BenchmarkReplayCollection, which measures the full small simulated collector plus normalization and analysis path for the internet-sensitive S3 scenario. Its value must be taken from the completed GitHub Actions log; no value is invented here until that run completes.

The required larger benchmark plan is to generate deterministic graphs of 10,000 nodes/50,000 edges and 50,000 nodes/250,000 edges, then measure persistence, loading, filtered graph queries, rule evaluation, path traversal, blast-radius traversal, API latency, memory, and SQLite size. Those measurements have not yet been executed.

A benchmark record must include command, hardware, Go version, dataset seed, and raw output. Distributed infrastructure is not an acceptable substitute for fixing accidental quadratic behavior.
