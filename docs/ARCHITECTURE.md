# Architecture

AegisGraph is a single-process service. The Go binary owns AWS/replay adapters, shared collectors, deterministic analysis, HTTP APIs, and SQLite. The compiled React application is static content served by the binary in packaged deployments.

## Source boundary

Official AWS SDK clients and replay clients are adapters only. Both implement the same narrow ReplayClient contract:

- caller identity and region discovery;
- paginated regional EC2/network responses;
- paginated IAM, S3, RDS, and Lambda responses.

Collectors do not receive AWS SDK response types. They consume provider-neutral replay-domain response objects and produce one normalized Snapshot. There is exactly one collector implementation per supported service.

## Pipeline

```
adapter -> service-shaped page -> shared collector/normalizer
        -> typed nodes/edges -> deterministic analysis
        -> findings/paths/blast-radius -> SQLite -> API/UI
```

The live path is LIVE_AWS_IMPLEMENTED and LIVE_AWS_UNVERIFIED until an authorized AWS account is used. Replay is REPLAY_VERIFIED.

## Persistence and lifecycle

SQLite uses foreign keys, WAL mode, bounded API queries, transactions, and indexes on node types/keys, edge endpoints, findings, and path scores. Scans are append-only records with source environment, account, timestamps, and coverage. A COMPLETE scan transaction replaces current materialized nodes, edges, findings, and paths. PARTIAL, FAILED, RUNNING, and PENDING scans record coverage but do not retire current resources or resolve current findings. Finding history tracks stable rule/resource keys, first/last seen, and resolved state.

## Analysis

Network exposure requires supported addressing, an Internet Gateway route, and an open supported TCP rule. IAM matching supports normalized Allow/Deny action/resource pairs, wildcard matching, and explicit Deny precedence. Unsupported semantics remain partial/unknown. Attack paths and blast radius traverse only allowlisted security transitions; containment is never attacker capability.

## Concurrency and scaling

The current service uses one SQLite connection and no background goroutines. Live AWS calls use context-aware SDK clients and SDK pagination. Future bounded worker pools and PostgreSQL are measurement-driven options, not required infrastructure.

## Failure model

Service errors are attributed to service/region coverage. AccessDenied, throttling, transient failures, missing fields, pagination cycles, and unsupported semantics cannot silently become complete or safe conclusions.
