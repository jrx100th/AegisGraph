# AegisGraph progress

## CURRENT STATE

Repository: jrx100th/AegisGraph. Branch: main. Task 3 core implementation is complete and the latest hosted verification is green at commit 19333aae810307f0f81000e663aec18d78b51b53.

## COMPLETED

- One shared narrow ReplayClient contract feeds both in-process/JSON replay and official AWS SDK adapters.
- Official read-only adapters exist for STS, EC2/VPC networking, IAM roles/users/policies/instance profiles, S3 bucket metadata, RDS instances, and Lambda functions.
- URL-decoded IAM policy retrieval, managed default versions, inline policies, public-state uncertainty, RDS network prerequisites, Lambda role links, and attached-user policies are represented.
- Append-only scan history records source/account/status/timestamps/coverage. Complete scans replace current materialized state; partial/failed scans preserve current resources and findings.
- Stable finding history keys track first/last seen and resolved state.
- Versioned JSON replay loader with strict schema validation, 8 MiB size bound, example fixtures, and API endpoint.
- Bounded capability-only blast-radius traversal with cycle prevention and evidence.
- Replay corpus expanded to 80 scenarios, with explicit 20 supported-risk and 20 false-positive suites.
- Adapter equivalence, JSON replay, lifecycle, and blast-radius tests added.
- CI now verifies startup through health/demo/findings/path requests and builds the Docker image.

## IN PROGRESS

- Final documentation-only CI rerun after this progress update.
- Manual live AWS validation remains intentionally unavailable in this session.
- Broader persistence/API measurements remain future work.

## TEST STATUS

Hosted CI run 34051561436, job 101535995306, passed Go formatting, go vet, go test ./..., go test -race ./..., frontend npm install/build, backend build, application startup smoke, and Docker build. No local Go/Docker/SQLite CLI was available in this Work container.

## BENCHMARK STATUS

Hosted run 34051561436 on linux/amd64 AMD EPYC 7763 measured BenchmarkDemoSnapshot 3,783 ns/op; BenchmarkNetworkReachability 126.3 ns/op; BenchmarkReplayCollection 7,556 ns/op; and BenchmarkBlastRadius10kNodes50kEdges 2,591,327 ns/op with 874,621 B/op and 111 allocs/op. The scale case uses 10,000 nodes and 50,006 edges. These are synthetic regression measurements, not production guarantees.

## KNOWN BUGS

No known failing tests or build defects remain in the verified replay/live-adapter path. Live AWS behavior has not been exercised against an AWS account.

## KNOWN LIMITATIONS

- LIVE_AWS_IMPLEMENTED and LIVE_AWS_UNVERIFIED are the correct live labels; LIVE_AWS_VALIDATED is not claimed.
- Unsupported IAM Conditions, permission boundaries, session policies, SCPs, resource policies, complete cross-account authorization, and service-specific semantics remain partial/unknown.
- Network ACLs, IPv6, NAT, load balancers, transit gateways, private routing, and service endpoints remain limited/unsupported.
- Sanitized fixtures are synthetic examples; no captured real-AWS fixture is included.
- Large persistence/API/50k graph measurements remain outstanding.

## IMPORTANT DECISIONS

- Preserve one collector implementation per supported service and keep adapters narrow.
- Treat partial/failed scans as non-destructive observations.
- Prefer UNKNOWN/PARTIAL over unsafe IAM/network certainty.
- Keep AWS-first scope and defer other providers/runtime AI.

## NEXT HIGHEST PRIORITIES

1. Exercise the live adapter against an authorized, read-only AWS account and label only the actually validated coverage.
2. Add 50k-node/250k-edge persistence and API measurements.
3. Add richer historical finding/scan API views.
4. Expand conservative handling for NACL/IPv6/load balancer semantics.
5. Revisit collector concurrency only after measured account-scale bottlenecks.
