# AWS Simulation Lab

The simulation lab is an offline verification subsystem for the collector pipeline. It is designed to exercise the same replay-facing collector contract, normalization, SQLite persistence, graph analysis, findings, and attack-path generation used by the application. It does not create findings or paths directly.

## Execution flow

```
SimulatedAWS responses
  -> ReplayClient collector
  -> normalized Snapshot
  -> analyzeSnapshot
  -> SQLite persistence
  -> API/UI consumers
```

`SimulatedAWS` implements the narrow `ReplayClient` interface in `cmd/aegisgraph/replay.go`. The simulator preserves page tokens and can return AccessDenied, throttling, transient, malformed, empty, incomplete, and out-of-order responses. The collector records service/region coverage as COMPLETE or PARTIAL and marks a scan FAILED when caller identity cannot be obtained.

Run the complete replay corpus with:

    go test ./cmd/aegisgraph -run Replay -count=1

Run all Go verification, including the replay corpus:

    go test ./...

The current catalog contains 49 deterministic scenarios covering network near-misses, IAM policy/trust cases, compound paths, lifecycle behavior, pagination, and service failures. The tests include explicit suites for 10 supported-risk cases and 13 false-positive traps.

## Scenario categories

- Network: private workloads, public SSH/RDP/database exposure, missing routes, wrong ports, restricted CIDRs, missing security-group data, IPv6, and NACL uncertainty.
- IAM: Allow, Deny, wildcard matching, inline/managed policy representation, trust, malformed documents, unsupported conditions, permission-boundary/SCP uncertainty, and cycles.
- Compound risk: public workloads with role access to sensitive resources, denied access, role-chain fixtures, and bounded cycle cases.
- Lifecycle/failure: initial and repeated scans, complete disappearance, partial scans, region/service AccessDenied, throttling, transient errors, missing identity metadata, and empty pages.

## Truth boundary

Simulation proves deterministic behavior for the modeled responses and protects against regressions in the collector-to-analysis pipeline. It is not equivalent to live AWS validation. It does not establish complete AWS API coverage, exact IAM parity, or provider network-semantic correctness. The live AWS collector is still a partial STS/EC2 implementation; IAM, S3, RDS, and Lambda live collection remain follow-up work.

Do not add a scenario that inserts a normalized node, finding, or path directly into storage. Add raw service responses and assert semantic outputs instead.
