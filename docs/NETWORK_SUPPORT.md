# Network support

## Supported reasoning

The deterministic model can conclude supported Internet reachability only when the modeled prerequisites are present:

1. A public address exists where the resource requires one.
2. A verified subnet route reaches an Internet Gateway.
3. A Security Group ingress rule permits a supported TCP port from 0.0.0.0/0.

The current exposure ports are TCP/22 (SSH), TCP/3389 (RDP), TCP/3306 (MySQL), and TCP/5432 (PostgreSQL). Port-range containment is evaluated conservatively. The replay collector can normalize EC2 and RDS-shaped responses into this model.

## UNKNOWN and limitations

Network ACLs, load balancers, transit gateways, IPv6, NAT, private routing, service endpoints, and provider-specific forwarding semantics are not implemented. A real collector must return UNKNOWN/PARTIAL when one of these can materially change the conclusion. The replay corpus includes IPv6 and NACL uncertainty cases.

A public IP alone is never treated as proof of Internet reachability. The live AWS collector currently covers only the existing partial EC2 inventory path; live RDS exposure collection is not yet implemented.
