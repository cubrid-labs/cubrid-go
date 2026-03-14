# Agent Playbook

## Source Of Truth
- `README.md` for public usage, DSN format, and feature expectations.
- `PRD.md` for architecture and roadmap intent.
- `CONTRIBUTING.md` and `Makefile` for day-to-day development commands.

## Repository Map
- Root Go packages provide the driver implementation.
- `docs/` stores supplementary documentation.
- Tests in the module root cover unit and integration behavior.

## Change Workflow
1. Decide whether the change affects the core driver, GORM support, or docs only.
2. Maintain `database/sql` behavior and error semantics unless a contract change is intentional.
3. Update README examples when DSN, transactions, or type mappings change.
4. Run integration tests when touching live-DB paths and CUBRID is available.

## Validation
- `go test ./...`
- `make vet`
- `make test-cov`
- `make integration` when integration paths are touched
