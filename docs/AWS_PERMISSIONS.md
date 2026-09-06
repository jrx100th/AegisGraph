# AWS permissions

The current partial collector uses the normal AWS SDK for Go credential chain and read-only calls only. It currently calls STS caller identity and EC2 regional inventory APIs. Before enabling additional services, each collector must preserve per-service coverage and partial-scan semantics.

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

The current binary performs the STS and EC2 calls needed for the partial inventory described above. IAM, S3, RDS, and Lambda permissions remain design targets until their collectors are implemented and verified.
