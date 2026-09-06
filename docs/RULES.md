# Rules

The v0.1 demo contains:

- AG-NET-001: Public SSH exposure, HIGH.
- AG-NET-002: Public RDP exposure, HIGH.
- AG-IAM-001: Broad administrative IAM privilege, CRITICAL.
- AG-COMB-002: Internet-exposed workload with supported access to sensitive resource, CRITICAL.

Every finding includes a stable rule ID, affected node, rationale, evidence, remediation, and integer risk score. Rules are produced from the snapshot engine, not hard-coded into the frontend.
