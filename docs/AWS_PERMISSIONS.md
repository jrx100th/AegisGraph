# AWS permissions

Live AWS collection is not enabled in v0.1. Before enabling it, collectors should use the normal AWS SDK for Go credential chain and read-only calls only.

The intended initial permission families are:

- sts:GetCallerIdentity
- ec2:DescribeRegions
- ec2:DescribeInstances
- ec2:DescribeAddresses
- ec2:DescribeVpcs
- ec2:DescribeSubnets
- ec2:DescribeRouteTables
- ec2:DescribeInternetGateways
- ec2:DescribeSecurityGroups
- iam:ListUsers
- iam:ListRoles
- iam:ListPolicies
- iam:ListInstanceProfiles
- iam:ListRolePolicies
- iam:ListUserPolicies
- s3:ListAllMyBuckets
- rds:DescribeDBInstances
- lambda:ListFunctions

These are design targets, not a claim that the current binary performs them.
