# Release checklist

Before calling a release usable, verify in a tool-equipped environment:

- gofmt, go vet, go test ./..., and go test -race ./...
- npm install and npm run build
- Docker build and container startup
- demo load for secure, exposed, attack-path, and false-positive-trap
- SQLite persistence and rescan behavior
- API response bounds and security headers
- secret scanning
- independent evidence reconstruction
- real benchmarks at both target graph sizes

The current repository intentionally does not mark these checks complete because the Work container lacks Go, Docker, and SQLite CLI tooling.
