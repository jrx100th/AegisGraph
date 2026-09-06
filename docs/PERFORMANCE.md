# Performance

Hosted GitHub Actions run 34051561436, job 101535995306, executed:

    go test -run "^$" -bench "^Benchmark" -benchmem ./...

Environment: linux/amd64, AMD EPYC 7763 64-Core Processor, Go 1.24 toolchain.

Measured output:

- BenchmarkDemoSnapshot: 3,783 ns/op, 5,936 B/op, 39 allocs/op.
- BenchmarkNetworkReachability: 126.3 ns/op, 112 B/op, 3 allocs/op.
- BenchmarkReplayCollection: 7,556 ns/op, 8,399 B/op, 69 allocs/op.
- BenchmarkBlastRadius10kNodes50kEdges: 2,591,327 ns/op, 874,621 B/op, 111 allocs/op.

The scale benchmark constructs 10,000 nodes and 50,006 edges; the benchmark name records the intended 50,000-edge workload plus six deterministic reachable edges. These are hosted synthetic measurements, not production latency or memory guarantees. Persistence, API latency, 50k/250k graph loading, and full scan throughput remain future measurements.
