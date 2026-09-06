# IAM support

## Deterministic runtime subset

The current analysis subset supports:

- identity-policy Allow and Deny records represented by normalized action/resource pairs;
- exact and trailing-wildcard action matching;
- exact and trailing-wildcard resource matching;
- explicit supported Deny overriding Allow;
- normalized role attachment and trust edges as evidence-bearing graph relationships.

The replay lab represents both inline role policy documents and attached managed-policy documents, and includes malformed, unsupported-condition, cross-account, permission-boundary, SCP, and trust-cycle inputs. Those cases are marked PARTIAL/UNKNOWN rather than being converted into Allow.

## Not implemented

Malformed policy semantics, Conditions, permission boundaries, session policies, SCPs, resource policies, service-specific authorization, and complete cross-account evaluation are not implemented in the live runtime. Full AssumeRole authorization is not claimed; trust fixtures are currently used to test normalization and uncertainty handling.

The live AWS collector does not yet collect IAM users, roles, managed policies, inline policies, or instance profiles. The replay collector models their response shape for offline verification only.

This is not AWS IAM simulator parity. Unsupported semantics must never silently become ALLOW.
