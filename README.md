# AegisGraph

AegisGraph is an open-source, deterministic cloud-security graph for authorized environments. It normalizes cloud assets into typed relationships, explains supported security conclusions with evidence, and keeps the runtime free of AI services.

This repository currently contains a verified v0.1 demo slice plus an offline AWS collector-replay laboratory:

- Go HTTP API with SQLite persistence and schema migrations.
- React + TypeScript investigation console.
- Secure, exposed, attack-path, and false-positive-trap synthetic environments.
- Conservative Internet exposure checks requiring supported addressing, Internet Gateway routing, and open inbound TCP rules.
- Bounded Internet-to-workload-to-role-to-sensitive-resource path generation.
- Explicit-deny-overrides-allow behavior for the implemented IAM subset.
- Transparent finding evidence, remediation text, and integer risk scores.
- A 49-scenario simulated AWS corpus that exercises pagination, partial coverage, lifecycle, IAM/network near-misses, and service failures through a collector-facing interface.

## Quick start

Requirements: Go 1.22 or newer, Node 20 or newer, and npm.

Build the frontend:

    cd frontend
    npm install
    npm run build
    cd ..

Run the API:

    go run ./cmd/aegisgraph -static frontend/dist

Open http://localhost:8080 and select Load demo. No AWS credentials or Internet connection are needed after dependencies are available.

The backend can also be exercised directly:

    curl -X POST 'http://localhost:8080/api/demo/load?environment=attack-path'
    curl http://localhost:8080/api/stats
    curl http://localhost:8080/api/findings
    curl http://localhost:8080/api/attack-paths

Supported demo environments are secure, exposed, attack-path, and false-positive-trap.

## Offline AWS simulation

Run the collector replay corpus:

    go test ./cmd/aegisgraph -run Replay -count=1

The simulator enters through the replay collector contract and then uses the ordinary normalization, analysis, persistence, findings, and attack-path code. It never inserts UI-only findings or hard-coded paths. See docs/SIMULATION.md and docs/AWS_REPLAY.md.

Simulation establishes deterministic behavior for modeled responses; it is not proof of complete real-world AWS behavior or live IAM/network parity.

## AWS status

The current live endpoint at POST /api/scans/aws is a bounded partial AWS inventory collector. It uses the normal AWS SDK credential chain and read-only STS/EC2 calls for caller identity, enabled regions, VPCs, subnets, route tables, Internet Gateways, security groups, and instances. It reports PARTIAL coverage and does not yet provide live IAM, S3, RDS, or Lambda collection. No credentials are persisted.

## Truthfulness boundary

This is not full AWS IAM equivalence, a vulnerability scanner, an exploit tool, or a Wiz replacement. Unsupported semantics must remain UNKNOWN. See docs/LIMITATIONS.md.

## Verification

    go test ./...
    go test -race ./...
    cd frontend && npm test && npm run build

The exact evidence status is maintained in PROGRESS.md and docs/VERIFICATION.md. Do not treat unexecuted local commands as passing.

## License

Apache-2.0. Contributions should preserve evidence, deterministic behavior, explicit uncertainty, and credential safety.
