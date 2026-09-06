# AegisGraph

AegisGraph is an open-source, deterministic cloud-security graph for authorized environments. It normalizes cloud assets into typed relationships, explains supported conclusions with evidence, and has no runtime AI dependency.

The current release is an AWS-first demo and replay-verified foundation:

- Go HTTP API, SQLite persistence, schema migrations, and React/TypeScript console.
- Secure, exposed, attack-path, and false-positive-trap demo environments.
- One shared collector pipeline for live AWS adapters and offline replay clients.
- Conservative Internet exposure reasoning requiring supported addressing, Internet Gateway routing, and an open supported TCP rule.
- Bounded IAM matching with explicit Deny precedence and uncertainty for unsupported conditions.
- Deterministic findings, attack paths, blast-radius traversal, evidence, remediation, and integer risk values.
- 80 deterministic replay scenarios, including pagination, service errors, lifecycle, IAM/network near-misses, S3/RDS/Lambda cases, and adapter equivalence.

## Quick start

Requirements: Go 1.24 or newer, Node 20 or newer, and npm.

    cd frontend
    npm install
    npm run build
    cd ..
    go run ./cmd/aegisgraph -static frontend/dist

Open http://localhost:8080 and select Load demo. No AWS credentials or network are needed after dependencies are available.

Useful API calls:

    curl -X POST 'http://localhost:8080/api/demo/load?environment=attack-path'
    curl http://localhost:8080/api/stats
    curl http://localhost:8080/api/findings
    curl http://localhost:8080/api/attack-paths
    curl http://localhost:8080/api/blast-radius/<node-key>

Supported demo environments are secure, exposed, attack-path, and false-positive-trap.

## Offline AWS replay

Run the deterministic collector corpus:

    go test ./cmd/aegisgraph -run 'Replay|JSON|BlastRadius|Adapter' -count=1

Load a sanitized JSON fixture through the same collector and analysis pipeline:

    curl -X POST http://localhost:8080/api/replay/load       --data-binary @fixtures/replay/attack-path.json       -H 'Content-Type: application/json'

Fixture format and sanitization guidance are in docs/AWS_REPLAY.md. Replay proves modeled behavior and collector robustness; it is not proof of complete real-world AWS behavior or IAM/network parity.

## AWS mode

POST /api/scans/aws uses the normal AWS SDK credential chain and the shared collectors. Implemented read-only adapters cover STS, EC2/VPC networking, IAM roles/users/policies/instance profiles, S3 bucket metadata, RDS instances, and Lambda functions. Service and region coverage are retained in the scan record.

Current capability labels are:

- DEMO_VERIFIED: synthetic demo pipeline.
- REPLAY_VERIFIED: deterministic simulator/fixture pipeline.
- LIVE_AWS_IMPLEMENTED: official SDK adapters and shared collectors are present.
- LIVE_AWS_UNVERIFIED: no live AWS account was used in this work session.
- PARTIAL_SUPPORT / UNKNOWN: a conclusion depends on unavailable or unsupported semantics.
- UNSUPPORTED: not evaluated.

No credentials are persisted or returned. No mutation-capable AWS calls are used. Do not label the live path LIVE_AWS_VALIDATED until it has been exercised against an authorized AWS account.

## Truthfulness boundary

This is not full AWS IAM equivalence, a vulnerability scanner, an exploit tool, or a replacement claim for a commercial platform. Conditions, permission boundaries, session policies, SCPs, resource policies, complete cross-account authorization, NACLs, IPv6, NAT, load balancers, and service-specific authorization remain limited or unsupported. See docs/LIMITATIONS.md.

## Verification

    go test ./...
    go test -race ./...
    cd frontend && npm test && npm run build

GitHub Actions runs the Go formatter, vet, tests, race suite, benchmarks, and frontend build without AWS credentials. Exact evidence and remaining gaps are maintained in PROGRESS.md and docs/VERIFICATION.md.

## License

Apache-2.0. Contributions should preserve evidence, deterministic behavior, explicit uncertainty, and credential safety.
