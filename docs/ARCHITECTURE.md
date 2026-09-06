# Architecture

AegisGraph is a single-process service. The Go binary owns the API, deterministic analysis, and SQLite database. The compiled React application is static content served by the same binary in packaged deployments.

## Boundary

The frontend only calls bounded JSON endpoints. It never receives credentials. The backend owns discovery, normalization, graph persistence, evidence, rule evaluation, path computation, and risk calculation.

## Persistence

SQLite uses foreign keys, WAL mode, transactions, stable node keys, and indexes on node type/key, edge endpoints, finding severity, and path score. A scan replacement transaction deletes prior materialized results before inserting one complete synthetic snapshot. Live scans must evolve this into per-scan lifecycle retention with explicit stale/deleted states.

## Domain and graph

Nodes have canonical keys, provider/account/region identity, type, display name, normalized properties, and provenance through scan-linked records. Edges are typed and carry evidence. The runtime currently reads the relational graph directly for API output; a compact adjacency structure should be introduced when traversal scale is measured.

## Analysis

The demo pipeline is:

fixture -> typed nodes/edges -> SQLite -> network evaluation -> bounded IAM matching -> findings -> allowed security transitions -> attack paths/blast radius -> API/UI.

The only currently supported transitions are Internet to verified exposed workload, workload to attached role, and supported role access to a marked sensitive resource. Containment never creates attacker capability.

## Failure model

Missing or unsupported prerequisites produce no positive security conclusion. Live collection must record service coverage as COMPLETE, PARTIAL, FAILED, or NOT_ATTEMPTED and must never label a partial account as complete.

## Concurrency and scaling

The current demo uses one SQLite connection and no background goroutines. Future collectors should use context cancellation and bounded workers. PostgreSQL is a future adapter only after measurements show SQLite is insufficient.

## Future modules

The next refactor should separate internal/domain, internal/store, internal/engine, internal/aws, and internal/http packages without changing the public semantics.
