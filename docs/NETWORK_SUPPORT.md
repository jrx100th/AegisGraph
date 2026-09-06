# Network support

## Supported reasoning

The shared model can conclude supported Internet reachability only when all modeled prerequisites are present:

1. A public address exists where the resource requires one.
2. The subnet has a verified route to an attached Internet Gateway.
3. A Security Group permits a supported TCP port from 0.0.0.0/0.

Supported ports are TCP/22 (SSH), TCP/3389 (RDP), TCP/3306 (MySQL), and TCP/5432 (PostgreSQL). Multiple Security Groups and port ranges are evaluated conservatively. EC2 and RDS responses can enter through the live adapter or replay with the same normalizer and reasoning.

## Uncertainty

A public address alone is never proof of reachability. Missing route, subnet, or Security Group evidence prevents a positive conclusion. RDS PubliclyAccessible is not enough without network evidence. IPv6, NACLs, load balancers, transit gateways, NAT, private routing, service endpoints, and provider forwarding semantics are not fully implemented; materially incomplete cases remain UNKNOWN/PARTIAL.

Status labels are REPLAY_VERIFIED, LIVE_AWS_IMPLEMENTED, and LIVE_AWS_UNVERIFIED. No live AWS account was used for this task.
