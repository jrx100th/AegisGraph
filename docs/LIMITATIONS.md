# Limitations

This release is a deterministic demo and architecture foundation, not a complete AWS security product.

The most important limitation is that live AWS discovery is not yet enabled. The current endpoint returns NOT_IMPLEMENTED and never accepts credentials.

IAM support is intentionally narrow. Network semantics are intentionally narrow. Resource policy, condition, SCP, permission-boundary, session-policy, NACL, load-balancer, IPv6, NAT, and cross-account semantics are not implemented.

The scan replacement transaction is suitable for the single demo snapshot but needs a richer lifecycle for production cloud scans, including scan IDs, historical views, stale/deleted resources, and partial-service visibility.

The frontend is a polished investigation console slice, not the complete required asset detail, scan status, settings, or large-graph interaction suite.

No performance number is published until the benchmark harness runs in an environment with Go.
