# Verification record

## Pass 1: implementation checks

Automated tests cover:

- Public address without an Internet Gateway route.
- Correct route plus an open supported port.
- Wrong port rejection.
- Explicit Deny overriding wildcard Allow.
- False-positive-trap path suppression.
- Attack-path shape and per-transition evidence.

The tests are committed in cmd/aegisgraph/main_test.go.

## Pass 2: adversarial checks

The fixtures intentionally include:

- A public IP without a route.
- A restricted private CIDR.
- A wrong-port near miss.
- A contradictory IAM Allow/Deny policy.
- A bounded multi-step path.

The current test source documents expected outcomes. Additional malformed-policy and cyclic-role tests are required when the live IAM parser is added.

## Pass 3: independent evidence reconstruction

For the attack-path fixture, the conclusion can be reconstructed manually from:

1. external:internet -> aws:ec2:i-demo, supported by public IP, IGW route, and open TCP/22.
2. aws:ec2:i-demo -> aws:role:demo-role, supported by the RUNS_AS fixture edge.
3. aws:role:demo-role -> aws:s3:demo-sensitive, supported by the normalized s3:GetObject Allow and the sensitive node marker.

The false-positive fixture has the same public address and ingress but lacks the IGW route, so no critical path is generated.

## Environment limitation

This Work container does not have Go, Docker, or SQLite command-line tooling installed. Therefore local test, race, benchmark, Docker, and clean-install results are not claimed. GitHub Actions is configured to run the backend and frontend checks in a tool-equipped environment.
