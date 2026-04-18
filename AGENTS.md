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

## Development Workflow (cubrid-labs org standard)

All non-trivial work across cubrid-labs repositories MUST follow this 4-phase cycle:

1. **Oracle Design Review** — Consult Oracle before implementation to validate architecture, API surface, and approach. Raise concerns early.
2. **Implementation** — Build the feature/fix with tests. Follow existing codebase patterns.
3. **Documentation Update** — Update ALL affected docs (README, CHANGELOG, ROADMAP, API docs, SUPPORT_MATRIX, PRD, etc.) in the same PR or as an immediate follow-up. Code without doc updates is incomplete.
4. **Oracle Post-Implementation Review** — Consult Oracle to review the completed work for correctness, edge cases, and consistency before merging.

Skipping any phase requires explicit justification. Trivial changes (typos, single-line fixes) may skip phases 1 and 4.

## Validation
- `go test ./...`
- `make vet`
- `make test-cov`
