# Contributing

Keep changes small, deterministic, and evidence-backed.

Before opening a pull request:

1. Read docs/ARCHITECTURE.md and docs/LIMITATIONS.md.
2. Add positive, negative, adversarial, and regression tests for security logic.
3. Run gofmt, go vet, go test ./..., and the frontend production build.
4. Never include credentials, cloud exports containing secrets, or generated database files.
5. Update PROGRESS.md and docs/VERIFICATION.md when behavior or support claims change.

A pull request must state what is supported, what remains UNKNOWN, and how the conclusion was independently checked.
