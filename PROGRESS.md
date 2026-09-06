# AegisGraph progress

## CURRENT STATE

Repository: jrx100th/AegisGraph. Branch: main. Current milestone: v0.1 deterministic demo slice.

## COMPLETED

- Go HTTP service and SQLite schema/migrations.
- Transactional replacement of the current synthetic scan.
- Typed nodes and edges with evidence.
- Secure, exposed, attack-path, and false-positive-trap fixtures.
- Network prerequisite evaluation and explicit-deny IAM subset.
- Findings, bounded attack paths, blast-radius endpoint, and integer risk values.
- React/TypeScript dashboard with graph, findings, and path panels.
- README, security policy, license, architecture/support/verification documents.
- Dockerfile and CI workflow.

## IN PROGRESS

- Live AWS read-only discovery and provider normalization.
- Wider IAM semantics, resource policies, conditions, and cross-account support.
- More complete scan lifecycle and stale-resource retention policy.

## TEST STATUS

The repository includes Go adversarial tests. Local execution in the Work container was blocked because Go and SQLite tooling are not installed. CI is configured to execute the commands on an official Go environment. No passing result is claimed locally.

## BENCHMARK STATUS

No benchmark number is claimed yet. The benchmark harness remains a next priority.

## KNOWN BUGS

- The AWS endpoint is intentionally not implemented.
- The current UI is a v0.1 desktop-oriented console and does not yet provide full asset-detail routing.

## KNOWN LIMITATIONS

See docs/LIMITATIONS.md. Unsupported semantics must not be interpreted as safe.

## IMPORTANT DECISIONS

- Keep Go + SQLite as the target runtime architecture.
- Use a custom bounded SVG graph in v0.1 to avoid an unmeasured graph dependency; evaluate Cytoscape.js or Sigma.js before larger graphs.
- Persist the demo through the same database and engine path as future cloud scans.
- Prefer explicit NOT_IMPLEMENTED/UNKNOWN over fabricated AWS conclusions.

## NEXT HIGHEST PRIORITIES

1. Add AWS SDK collectors with pagination, coverage states, cancellation, and read-only tests.
2. Split domain, storage, engine, and HTTP packages before broadening coverage.
3. Add benchmark generation and pprof evidence.
4. Add clean-install and container verification from an environment with Go and Docker.
