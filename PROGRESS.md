# AegisGraph progress

## CURRENT STATE

Repository: jrx100th/AegisGraph. Branch: main. Current milestone: v0.1 deterministic demo slice plus offline AWS collector replay laboratory. The live AWS endpoint remains a partial STS/EC2 inventory.

## COMPLETED

- Go HTTP service, SQLite schema/migrations, transactional persistence, and React/TypeScript console.
- Secure, exposed, attack-path, and false-positive-trap demo environments.
- Shared deterministic analysis for network findings, bounded attack paths, explicit supported IAM Deny precedence, evidence, remediation, and risk values.
- Narrow replay-facing AWS service interface and deterministic simulator for STS, EC2, IAM, S3, RDS, and Lambda-shaped responses.
- Collector-to-normalization-to-analysis replay tests with 49 scenarios, pagination, empty pages, AccessDenied, throttling/transient failures, malformed metadata, identity collision checks, and partial-scan retention.
- Offline scenario and replay documentation, CI execution, and a collector replay microbenchmark.

## IN PROGRESS

- Route the live official AWS SDK implementation through the replay-facing service interfaces.
- Implement live IAM users/roles/policies, instance profiles, S3, RDS, and Lambda collectors with service/region coverage.
- Expand lifecycle persistence from the replay ledger to historical scan IDs, stale-resource policy, and per-region visibility.
- Add sanitized JSON fixture loading and larger graph benchmark generation.

## TEST STATUS

GitHub Actions run 34030372218, job 101478659148, passed gofmt, go vet, Go tests, Go race tests, Go microbenchmarks, frontend install, and frontend production build. The Work container used for this session does not provide Go, Docker, or SQLite CLI tooling, so no local pass is claimed.

The replay suite explicitly verifies 10 supported-risk cases and 13 false-positive traps. Synthetic results validate modeled behavior only and do not replace live AWS validation.

## BENCHMARK STATUS

Run 34030372218 measured BenchmarkDemoSnapshot at 3,208 ns/op, BenchmarkNetworkReachability at 100.5 ns/op, and BenchmarkReplayCollection at 5,193 ns/op. No large-graph benchmark claim is made.

## KNOWN BUGS

- The live AWS collector is not yet using the replay-facing interfaces.
- The replay collector models IAM roles/policies, S3, RDS, and Lambda, but those service collectors are not yet wired to official live SDK clients.
- Scan history and stale/deleted-resource handling are stronger in replay tests than in the production endpoint.

## KNOWN LIMITATIONS

- Simulation is deterministic and reviewable but cannot prove complete AWS API, IAM, routing, or regional behavior.
- Conditions, permission boundaries, session policies, SCPs, resource policies, cross-account authorization, NACLs, IPv6, NAT, load balancers, and service-specific authorization are not fully implemented.
- Clean Docker execution and live AWS execution remain unverified in this Work container.

## IMPORTANT DECISIONS

- Keep Go + SQLite and use narrow service interfaces rather than wrapping the entire AWS SDK.
- Keep the simulator in-process and dependency-light.
- Exercise the real replay collector and analysis pipeline; never insert synthetic findings or paths directly.
- Prefer PARTIAL/UNKNOWN over unsafe certainty.
- Keep AWS-first scope; defer other providers and runtime AI.

## NEXT HIGHEST PRIORITIES

1. Adapt the live AWS SDK collector to the same narrow service interfaces.
2. Add live IAM/S3/RDS/Lambda collection and scan coverage/lifecycle persistence.
3. Add a sanitized JSON replay loader and verify it against captured, consented responses.
4. Run larger 10k/50k-node measurements and verify Docker/clean-install paths.
