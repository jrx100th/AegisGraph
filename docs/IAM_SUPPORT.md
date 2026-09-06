# IAM support

The shared collector and official AWS adapter represent IAM roles and trust principals, IAM users, instance profiles, inline role/user policies, attached role/user policies, managed-policy metadata and default-version documents, URL-decoded policy documents, Allow/Deny, and string/list Action and Resource values.

Matching supports exact values and trailing wildcards. A supported explicit Deny overrides Allow. Malformed policy documents, unsupported Conditions, unsupported trust Conditions, permission boundaries, SCPs, session policies, resource policies, and incomplete cross-account authorization remain PARTIAL/UNKNOWN; they are never silently converted to Allow.

REPLAY_VERIFIED covers deterministic modeled responses, including URL encoding and contradiction cases. LIVE_AWS_IMPLEMENTED means the SDK adapter exists; LIVE_AWS_UNVERIFIED remains the honest status until an authorized AWS account is used. This is not AWS IAM simulator parity or full AssumeRole authorization.
