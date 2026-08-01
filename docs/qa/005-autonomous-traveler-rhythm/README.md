# Autonomous traveler rhythm — verification index

This directory contains only redacted verification evidence: command names,
test names, exit status, timestamps, and screenshots without credentials.
Never store registration responses, passwords, Agent API keys, session IDs,
Set-Cookie values, raw database dumps, or HAR files here.

| Case group | Role | PRD acceptance criteria | Automated coverage | Latest command | Result |
| --- | --- | --- | --- | --- | --- |
| Registration, password-change boundary, owner-scoped snapshots | Agent + user | AC-004, AC-007–AC-011, AC-013, AC-015 | `tests/integration/hub_test.go` | `go test -count=1 -tags=integration ./tests/integration` | Pass — 2026-08-01 |
| Daily creation, durable checkpoints, return and album exactly-once behavior | System | AC-001–AC-003, AC-012, AC-014, AC-017 | `internal/journey/*_test.go` and `internal/journey/*_integration_test.go` | `go test -count=1 -tags=integration ./internal/journey` | Pass — 2026-08-01 |
| Browser projection, fault display, no local departure control | User | AC-005, AC-006, AC-014, AC-016 | `frontend/src/**/*.test.*`, `frontend/e2e/*.spec.ts` | `npm test -- --run && npm run lint && npm run build && npm run test:e2e` | Pass — 2026-08-01 |
| Go package regression suite | System | Supporting regression coverage | `./...` | `go test -count=1 ./...` | Pass — 2026-08-01 |

The `agent/`, `user/`, and `system/` folders are intentionally empty until a
staging or production run can provide redacted artefacts.  Local verification
is recorded in this index rather than copying potentially sensitive logs.

## Known release gates

- Production deployment requires HTTPS and a real email-verification delivery
  provider. The development `TRAVELINGHUB_AUTO_VERIFY_EMAIL=true` setting is
  not a production setting.
- A staging migration rehearsal, backup/restore drill, and HTTPS browser
  smoke test require deployment credentials and are therefore not represented
  by local test output.
