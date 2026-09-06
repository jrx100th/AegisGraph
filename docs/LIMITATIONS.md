# Limitations

This is an AWS-first deterministic demo and replay-verified foundation, not a complete cloud-security product.

## Live AWS coverage

Official AWS SDK adapters and shared collectors implement STS, EC2/VPC networking, IAM roles/users/policies/instance profiles, S3 bucket metadata, RDS instances, and Lambda functions. The honest status is LIVE_AWS_IMPLEMENTED plus LIVE_AWS_UNVERIFIED: no live AWS account was used in this work session. LIVE_AWS_VALIDATED is not claimed.

## Replay boundary

The offline corpus contains 80 named scenarios, 20 supported-risk cases, and 20 false-positive traps. The JSON loader accepts versioned sanitized fixtures and runs the same collector/normalization/analysis pipeline. Replay establishes deterministic modeled behavior, not complete AWS API, IAM, routing, or regional correctness.

## Security semantics

Conditions, permission boundaries, session policies, SCPs, resource policies, complete cross-account authorization, service-specific permissions, NACLs, load balancers, transit gateways, IPv6, NAT, private routing, and service endpoints are limited or unsupported. Unknown or partial evidence is not promoted to a positive conclusion.

## Lifecycle and scale

Complete scans replace the current materialized graph; partial/failed scans do not retire resources or resolve findings. Finding history has stable keys and resolved state. Large 10k/50k graph measurements and local Docker execution remain environment-dependent and are not claimed without recorded output.

## Product scope

The frontend is functional for demo/replay-backed investigation, but richer historical scan views, complete asset-detail workflows, and large-graph interaction remain future work.
