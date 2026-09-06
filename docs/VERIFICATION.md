# Verification record

## Pass 1: implementation checks

The committed Go tests cover the original demo and the offline replay pipeline:

- public address without an Internet Gateway route;
- correct route plus an open supported port;
- wrong port and restricted CIDR rejection;
- explicit Deny overriding wildcard Allow;
- 49 collector replay scenarios;
- pagination, empty pages, AccessDenied, throttling/transient errors, malformed metadata, duplicate-looking identities, and partial scan retention;
- 10 explicit supported-risk cases and 13 false-positive traps;
- normalized replay output persisted through the real SQLite persistence method.

The replay tests are in cmd/aegisgraph/replay_test.go. They compare semantic findings, paths, coverage, identity keys, and evidence rather than unstable full snapshots.

GitHub Actions run 34030372218, job 101478659148, passed gofmt, go vet, Go tests, Go race tests, Go microbenchmarks, frontend install, and frontend production build.

## Pass 2: adversarial checks

The corpus intentionally includes:

- public IP without a route;
- Internet route without a public address;
- wrong protocol/port and restricted source CIDR;
- missing route or Security Group data;
- unsupported IPv6/NACL information;
- contradictory IAM Allow/Deny;
- malformed policies, unsupported conditions, cross-account/permission-boundary/SCP uncertainty;
- cyclic trust;
- AccessDenied, throttling, transient failure, empty pages, and resource disappearance during complete versus partial scans.

The test suite prevents an unsupported compound path from being generated in the false-positive traps. A partial scan does not retire an existing resource in ReplayScanLedger.

## Pass 3: independent evidence reconstruction

For the replay attack-path fixture, the result can be reconstructed without trusting the path builder:

1. the raw regional response gives the workload a public address;
2. the raw subnet response marks a known Internet Gateway route;
3. the raw Security Group response permits TCP/22 from 0.0.0.0/0;
4. the raw instance response attaches demo-role;
5. the raw IAM response gives demo-role s3:GetObject on the sensitive bucket ARN;
6. the raw S3 response marks demo-sensitive as sensitive.

The resulting path is exactly Internet -> workload -> role -> sensitive resource, with one evidence record per transition. The denied-policy fixture has the same modeled exposure but an explicit supported Deny and produces no equivalent compound path.

## Truth boundary

These are deterministic synthetic/replay checks. They validate the modeled pipeline and adversarial behavior, not full real-world AWS correctness. The live AWS collector is still only partial STS/EC2, and no live AWS account was used in this verification.

The Work container does not have Go, Docker, or SQLite command-line tooling, so local execution is not claimed. Clean Docker and live AWS execution remain unverified.
