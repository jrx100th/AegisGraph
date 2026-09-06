# Performance

No benchmark values are published yet.

The required benchmark plan is to generate deterministic graphs of 10,000 nodes/50,000 edges and 50,000 nodes/250,000 edges, then measure persistence, loading, filtered graph queries, rule evaluation, path traversal, blast-radius traversal, API latency, memory, and SQLite size.

A benchmark must run in a Go-equipped environment and record command, hardware, Go version, dataset seed, and raw output. Distributed infrastructure is not an acceptable substitute for fixing accidental quadratic behavior.
