# IAM support

The current subset supports deterministic matching of:

- Identity policy Allow and Deny records represented by normalized action/resource pairs.
- Exact and trailing-wildcard action matching.
- Exact and trailing-wildcard resource matching.
- Explicit supported Deny overriding Allow.

Malformed policies, conditions, permission boundaries, session policies, SCPs, resource policies, service-specific authorization, and cross-account evaluation are not implemented in the current runtime. They must produce UNKNOWN in live evaluation rather than silently becoming Allow.

This is not AWS IAM simulator parity.
