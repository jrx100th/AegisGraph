# AWS permissions

The live path uses the normal AWS SDK for Go credential provider chain. It performs read-only calls and does not persist credentials. Required calls depend on enabled regions and services:

- sts:GetCallerIdentity
- ec2:DescribeRegions, DescribeVpcs, DescribeSubnets, DescribeRouteTables, DescribeInternetGateways, DescribeSecurityGroups, DescribeInstances
- iam:ListUsers, ListRoles, ListPolicies, ListInstanceProfiles, ListRolePolicies, GetRolePolicy, ListAttachedRolePolicies, ListUserPolicies, GetUserPolicy, ListAttachedUserPolicies, GetPolicy, GetPolicyVersion
- s3:ListAllMyBuckets, GetBucketLocation, GetPublicAccessBlock, GetBucketPolicy, GetBucketEncryption, GetBucketTagging
- rds:DescribeDBInstances
- lambda:ListFunctions

Missing permissions are attributed to service or region and produce PARTIAL coverage. Exact service-specific authorization remains limited. The adapter is LIVE_AWS_IMPLEMENTED and LIVE_AWS_UNVERIFIED; no live account was used.
