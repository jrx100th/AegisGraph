# Limitations

This release is a deterministic demo and architecture foundation, not a complete AWS security product.

The current AWS endpoint provides partial read-only STS/EC2 inventory. It does not yet implement IAM, S3, RDS, Lambda, complete pagination across every service, or live finding/path analysis. It reports PARTIAL coverage and never persists credentials.

IAM support is intentionally narrow. Network semantics are intentionally narrow. Resource policy, condition, SCP, permission-boundary, session-policy, NACL, load-balancer, IPv6, NAT, and cross-account semantics are not implemented.

The scan replacement transaction is suitable for the single demo snapshot but needs a richer lifecycle for production cloud scans, including scan IDs, historical views, stale/deleted resources, and partial-service visibility.

The frontend is a polished investigation console slice, not the complete required asset detail, scan status, settings, or large-graph interaction suite.

No performance number is published until the benchmark harness runs in an environment with Go.
