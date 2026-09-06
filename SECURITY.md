# Security policy

AegisGraph is defensive software for environments the operator owns or is authorized to assess.

Do not report credentials, tokens, private keys, or live account data in an issue. For a suspected vulnerability, provide a minimal reproducible description through the repository security contact or a private security advisory when enabled.

Current security boundaries:

- The runtime has no AI dependency.
- The current demo build does not accept AWS credentials.
- No active exploitation, credential harvesting, arbitrary command execution, or destructive remediation is implemented.
- API responses are bounded and security headers are set by default.
- AWS collectors must use the normal read-only credential chain when implemented.
