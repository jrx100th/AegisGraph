# AWS Simulation Lab

The simulation lab is an offline verification subsystem. Raw simulated service responses enter the same narrow service interfaces consumed by the official AWS adapters. The shared collector performs normalization, graph construction, reasoning, findings, attack paths, and persistence-facing snapshot production.

## Flow

```
Simulated or replayed responses
  -> shared ReplayClient contract
  -> shared collectors and normalizers
  -> analyzeSnapshot
  -> SQLite/API/UI
```

SimulatedAWS preserves page tokens and can model normal responses, empty pages, pagination, AccessDenied, throttling, transient failures, malformed or incomplete data, multiple accounts/regions, and lifecycle changes.

Run the corpus:

    go test ./cmd/aegisgraph -run 'Replay|JSON|BlastRadius|Adapter' -count=1

The catalog contains 80 deterministic scenarios. The explicit suites include 20 supported-risk cases and 20 false-positive traps. The tests compare semantic findings, paths, coverage, identities, and evidence rather than unstable JSON ordering.

## Scenario categories

- Network: private workloads, SSH/RDP/database exposure, missing routes, wrong ports, restricted CIDRs, multiple groups, IPv6, and NACL uncertainty.
- IAM: Allow, Deny, wildcard matching, inline/managed representations, URL-encoded policy parsing, malformed documents, unsupported conditions, trust, account isolation, and cycles.
- S3/RDS/Lambda: public-state uncertainty, access-block signals, encryption metadata, RDS network prerequisites, execution-role relationships, and unsupported external exposure.
- Lifecycle/failure: complete versus partial disappearance, finding resolution, reappearance, service AccessDenied, throttling/transient errors, pagination, and missing identity metadata.
- Equivalence: different clients supplying the same logical pages and errors must produce the same normalized semantics.

Do not add a scenario that inserts a normalized node, finding, or path directly into storage. Add raw service-shaped responses and assert the result produced by the shared collector.

## Truth boundary

REPLAY_VERIFIED means deterministic modeled behavior passed the automated and adversarial suites. It is not LIVE_AWS_VALIDATED: no live AWS account was used here, and simulation does not establish complete AWS API behavior, exact IAM parity, or provider routing correctness.
