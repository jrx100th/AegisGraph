# Network support

## Supported

The current deterministic model can conclude Internet reachability only when all supported prerequisites are present:

1. A public address exists.
2. A verified subnet route reaches an Internet Gateway.
3. A Security Group ingress rule permits a supported TCP sensitive port from 0.0.0.0/0.

The model recognizes TCP/22 and TCP/3389 for exposure findings. Port range containment is evaluated conservatively.

## UNKNOWN and limitations

Network ACLs, load balancers, transit gateways, IPv6, NAT, private routing, service endpoints, and provider-specific forwarding semantics are not implemented. A real collector must return UNKNOWN when one of these can materially change the conclusion.

A public IP alone is never treated as proof of Internet reachability.
