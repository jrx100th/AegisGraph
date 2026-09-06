# AWS response replay

AegisGraph's current offline replay corpus is defined as deterministic Go values implementing the narrow `ReplayClient` interface. This keeps fixtures reviewable and avoids storing credentials or accidental live-account data.

## Replay contract

The simulator exposes these operations:

- caller identity and region discovery;
- paginated EC2 regional descriptions;
- paginated IAM pages;
- paginated S3 bucket pages;
- paginated RDS pages per region;
- paginated Lambda pages per region.

Each operation accepts `context.Context`, preserves page-token behavior, and can return a service-shaped error. The collector—not the test—normalizes the responses and invokes the ordinary analysis engine.

## Future sanitized fixture format

When real AWS responses are captured, contributors should convert them to a versioned fixture before committing:

```json
{
  "schema_version": 1,
  "source": "aws-sanitized",
  "account": "account-redacted",
  "regions": [
    {
      "name": "region-redacted",
      "ec2_pages": [
        {
          "next_token": null,
          "vpcs": [],
          "subnets": [],
          "security_groups": [],
          "instances": []
        }
      ],
      "rds_pages": [],
      "lambda_pages": []
    }
  ],
  "iam_pages": [],
  "s3_pages": []
}
```

Before committing a fixture, remove or replace account IDs, ARNs, resource names, IP addresses, tags, user data, policy identifiers, and any authentication material. Never commit access keys, session tokens, secret keys, cookies, or raw credential-provider output. Keep page boundaries and error responses intact because they are part of replay coverage.

A JSON loader is not yet part of the runtime. The current Go scenario catalog is the executable format; this schema documents the compatibility target for later sanitized-response ingestion.

## What replay can establish

Replay can establish deterministic normalization, identity-key construction, pagination handling, coverage state transitions, evidence preservation, and behavior for explicitly modeled IAM/network cases. It cannot establish that AWS returns all modeled fields in every service, that modeled fields have identical semantics in every region, or that the implementation matches the full AWS authorization and routing systems. Live AWS validation remains required.
