# Verification record

## Pass 1: implementation verification

The Go suite covers shared live/replay adapter equivalence; 80 replay scenarios; 20 supported-risk cases; 20 false-positive traps; EC2/RDS network prerequisites; IAM Allow/Deny, wildcard, URL-decoded documents, malformed documents, unsupported Conditions, trust cycles, and account/region isolation; S3 uncertainty; RDS/Lambda normalization; pagination; empty pages; AccessDenied; throttling/transient errors; missing fields; complete/partial persistence; finding resolution; bounded capability-only blast radius; and strict JSON fixture validation.

Tests compare semantic findings, paths, coverage, identities, and evidence rather than unstable ordering.

## Pass 2: adversarial verification

The corpus attempts false positives from public IPs without routes, routes without public addresses, wrong ports, restricted CIDRs, incomplete S3 state, RDS public flags without network evidence, Lambda existence, denied IAM access, unsupported Conditions, cycles, missing service data, and partial scans. Expected outcomes are no equivalent high-severity conclusion or explicit partial/unknown state.

## Pass 3: independent reconstruction

A representative replay path is reconstructed from raw regional, subnet, Security Group, instance, IAM, and S3 fields. Shared analysis emits Internet -> workload -> role -> sensitive resource with evidence for each transition. Blast radius traverses only RUNS_AS, CAN_ACCESS, and CAN_ASSUME edges; containment is excluded, cycles are visited once, and missing targets are uncertain.

These are DEMO_VERIFIED and REPLAY_VERIFIED results. The live path is LIVE_AWS_IMPLEMENTED and LIVE_AWS_UNVERIFIED; no live AWS account was used, so LIVE_AWS_VALIDATED is not claimed. Replay is not proof of complete AWS correctness.

GitHub Actions executes formatting, vet, tests, race tests, benchmarks, and frontend build. Local Go/Docker/SQLite CLI execution was unavailable in this Work container.
