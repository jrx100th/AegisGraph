# Graph model

The graph uses precise security relationships rather than generic connectivity.

Current types include AWS_ACCOUNT, VPC, SUBNET, EC2, IAM_ROLE, S3_BUCKET, and INTERNET.

Current edge types include CONTAINS, RUNS_AS, EXPOSED_TO, and CAN_ACCESS. Every security-relevant edge carries an evidence string. A graph edge without evidence is not eligible for an attack path.

Canonical identity must include provider, account, region where applicable, resource type, and native identifier. A native identifier alone is not globally unique.

Containment edges are inventory relationships only. They do not imply network reachability, permissions, exploitability, or attacker movement.
