# AGENTS.md

## Purpose
`cubrid-go` provides a pure Go CUBRID driver and related ecosystem integrations.

## Read First
- `README.md`
- `PRD.md`
- `CONTRIBUTING.md`
- `docs/agent-playbook.md`

## Working Rules
- Preserve `database/sql` semantics and DSN behavior unless the change explicitly updates those contracts.
- Keep documentation examples synchronized with the real driver behavior.
- Prefer compatibility-safe changes to exported types and behavior.
- Add tests for protocol, scanning, or transaction changes.

## Validation
- `go test ./...`
- `make vet`
- `make test-cov`
