# Limitations

This release is a deterministic demo and verification foundation, not a complete AWS security product.

## Live AWS coverage

The current AWS endpoint provides partial read-only STS/EC2 inventory. It does not yet route through the replay-facing interfaces or implement live IAM, S3, RDS, or Lambda collectors. It reports PARTIAL coverage and never persists credentials.

## Simulation boundary

The offline replay lab contains 49 code-defined scenarios for modeled STS, EC2, IAM, S3, RDS, and Lambda responses. It exercises the real replay collector, normalization, analysis, and SQLite persistence. It does not prove complete AWS API behavior, IAM parity, regional semantics, or network correctness. No live AWS account was used.

A versioned sanitized JSON format is documented in docs/AWS_REPLAY.md, but a JSON fixture loader and real-response corpus are not yet implemented. The current executable fixture format is Go.

## Security semantics

IAM support is intentionally narrow. Conditions, permission boundaries, session policies, SCPs, resource policies, service-specific authorization, full cross-account semantics, and full AssumeRole evaluation are not implemented. Network ACLs, load balancers, transit gateways, IPv6, NAT, private routing, and service endpoints are not implemented.

## Lifecycle and scale

The replay ledger protects against retiring resources after incomplete scans, but the production database still needs richer scan IDs, historical views, stale/deleted-resource policy, and per-region/service visibility. No 10k/50k large-graph benchmark has been executed.

## Product scope

The frontend is a polished investigation-console slice, not the complete required asset-detail, scan-status, settings, or large-graph interaction suite. Clean Docker execution and live AWS execution remain unverified in this Work container.
