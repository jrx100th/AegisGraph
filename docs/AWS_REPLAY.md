# AWS response replay

AegisGraph supports deterministic in-process simulation and a versioned sanitized JSON fixture format. Both feed the same narrow ReplayClient interface and shared collectors.

## JSON format

Schema version 1 uses maps from region names to page arrays:

```json
{
  "schema_version": 1,
  "source": "sanitized-example",
  "account": "111111111111",
  "regions": ["us-east-1"],
  "region_pages": {
    "us-east-1": [
      {
        "vpcs": [],
        "subnets": [],
        "security_groups": [],
        "instances": [],
        "rds": [],
        "lambda": [],
        "next_token": ""
      }
    ]
  },
  "iam_pages": [],
  "s3_pages": [],
  "rds_pages": {"us-east-1": []},
  "lambda_pages": {"us-east-1": []}
}
```

The loader rejects unsupported schema versions, duplicate or empty regions, unknown JSON fields, trailing JSON, and bodies larger than 8 MiB. Page next_token and error_code values are preserved so pagination and service failures remain testable.

Run a fixture through the API:

    curl -X POST http://localhost:8080/api/replay/load       --data-binary @fixtures/replay/attack-path.json       -H 'Content-Type: application/json'

Example fixtures are in fixtures/replay/. They are sanitized synthetic examples, not captured AWS responses.

## Sanitization guidance

Future contributors may convert consented AWS responses into this format with a deterministic replacement map:

1. Replace account IDs consistently.
2. Replace ARNs while preserving account, region, resource type, and relationships.
3. Replace names and public/private addresses consistently.
4. Remove user data, tags, policy identifiers, tokens, cookies, keys, and secret values unless structurally required and safely synthetic.
5. Preserve page boundaries, missing fields, and service errors.
6. Validate with LoadReplayFixture before committing.

No credentials or executable content belong in fixtures. Fixture loading is data-only and does not read arbitrary paths or invoke shell commands.

## What replay can establish

Replay can establish deterministic normalization, identity-key construction, pagination handling, coverage behavior, evidence preservation, lifecycle safety, and the modeled IAM/network conclusions. It cannot establish full AWS service behavior, exact IAM authorization parity, or live regional/network semantics. The status is REPLAY_VERIFIED; live validation remains a separate requirement.
