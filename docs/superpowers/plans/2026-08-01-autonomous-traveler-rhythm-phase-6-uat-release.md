# Autonomous Traveler Rhythm — Phase 6: End-to-end UAT and Release Readiness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Demonstrate that Agent and User flows meet every PRD acceptance criterion, document operational limits, and prepare a safe first deployment of the autonomous traveler rhythm.
**Architecture:** Verification is layered: unit tests prove clock and credential logic; integration tests prove PostgreSQL/Redis transactional behavior; Playwright proves browser behavior over HTTPS. Operations use one API process with its embedded bounded worker and PostgreSQL as the scheduling source of truth.
**Tech Stack:** Docker Compose, Go test, PostgreSQL 16, Redis 7, npm, Vitest, Playwright, HTTPS reverse proxy or local test certificate.

---

## Evidence locations and handling

Create the following directories with a short README explaining that evidence contains only redacted IDs, timestamps, command output, and screenshots:

- docs/qa/005-autonomous-traveler-rhythm/agent/
- docs/qa/005-autonomous-traveler-rhythm/user/
- docs/qa/005-autonomous-traveler-rhythm/system/

Never archive raw registration responses, Set-Cookie values, passwords, Agent API keys, full session IDs, database dumps, or browser HAR files that may contain them.

## Tasks

- [ ] **Create the verification harness and evidence index.** Add docs/qa/005-autonomous-traveler-rhythm/README.md with columns Case, role, AC, automated test path, command, redacted evidence file, owner, date, and result. Add an empty .gitkeep in each role directory. Add a root Makefile target named verify-autonomous-traveler that runs Go unit tests, Go integration tests, frontend unit/lint/build, and E2E in dependency order.

- [ ] **Automate Agent registration and credential cases.** Confirm AG-01 through AG-03 in tests/integration/identity_test.go: first registration is atomic; malformed/injected-failure operations leave no orphan records; concurrent and sequential duplicate email attempts disclose no IDs or credentials. Store only test names and redacted assertions in agent evidence.

- [ ] **Automate Agent feed/isolation cases.** Confirm AG-04 and AG-05 in tests/integration/agent_game_test.go: an API key sees only its own identity, snapshot, event cursor, and acknowledgement effects; cross-Agent cursor/resource attacks fail without altering the victim state.

- [ ] **Automate Agent offline lifecycle cases.** Confirm AG-06 and AG-07 in tests/integration/journey_worker_test.go: with no HTTP request, the worker creates the single 08:00 journey, survives duplicate cycles/restart, returns once, and keeps template/departure/return stable.

- [ ] **Automate user account and credential-boundary cases.** Confirm US-01 through US-03 in API tests plus Playwright: initial login forces password change; cookies have expected attributes; neither storage nor UI/network payload exposes Agent API keys; IDs injected by user A cannot select user B's traveler.

- [ ] **Automate visual state continuity cases.** Confirm US-04 through US-09 in frontend/e2e/journey.spec.ts with controlled server time: home-to-travelling, refresh/new-browser continuity, hidden in-flight return time, returned postcard exactly once, next-day 07:00 catch-up, and next-day 08:01 new departure with historic album retained. Capture screenshot paths in the evidence index.

- [ ] **Automate UI fault containment.** Confirm US-10 in Playwright by making the controlled API return a 500, expired session, and unknown template. Assert a human-readable retry/login/content error, preserved last valid screen where applicable, no unhandled exception, and no locally fabricated journey content.

- [ ] **Automate boundary and repository checks.** Add a script or Go test that fails if frontend/node_modules, frontend/dist, frontend/test-results, or frontend/.git exists. Confirm SY-01 by clean npm install/test/build/E2E. Confirm SY-02 with all temporal edge cases. Confirm SY-03 and SY-04 by negative API tests plus log capture that asserts credentials are absent.

- [ ] **Exercise migration upgrade and recovery.** Start a disposable database at migration 000001, apply all migrations once, restart the application, and verify existing legacy Agent rows remain readable while new registrations receive the full Web binding. Back up the database before migration in the release runbook. Record that rollback after user-created journey data is restore-from-backup, not an unsafe destructive down migration.

- [ ] **Complete production email verification readiness.** Select and configure the mail-delivery adapter and sending domain, send a disposable verification email in staging, verify expiration and single-use behavior, and prove production mode rejects login until verification completes. Do not permit a development bypass flag in the production deployment configuration.

- [ ] **Write operational runbook.** Add docs/operations/autonomous-traveler.md describing configuration, required HTTPS termination, database backup/restore, migration command, worker cadence/batch sizing, advisory-lock observability, health checks, alert signals, safe restart, and time-zone limitations. Include symptom-to-check mappings for missed journey, duplicate event, no session, and frontend cannot reach API.

- [ ] **Add observability without secret data.** Add only aggregate structured logs or metrics for worker cycle duration, candidate/departure/stage/return counts, lock contention, reconciliation errors, login success/failure class, and registration duplicate class. Test log redaction against generated credentials. Do not add individual email, key, password, session, or event payload values.

- [ ] **Perform deployment smoke test.** Deploy or stage behind HTTPS, register a disposable email, complete first-password change, load the frontend, verify the secure cookie is sent to same-origin API, and run a forced fixed-clock staging journey. Redact all artifacts before storing evidence.

- [ ] **Run final quality gate.** From a clean checkout, run make verify-autonomous-traveler, git diff --check, dependency vulnerability review appropriate to the project policy, and a manual review of changed migrations/session/cookie code. Resolve failures before proceeding.

- [ ] **Finalize the release record.** Fill the evidence index for AG-01 to AG-07, US-01 to US-10, and SY-01 to SY-04; link only redacted output/screenshot files. Update README.md with current API and startup instructions, tag or commit the release with a message such as docs: verify autonomous traveler rhythm.

## Release criteria

- [ ] Every PRD acceptance criterion has passing, current evidence or an explicitly approved manual verification record.
- [ ] A secure cookie works over deployed HTTPS; no development-only credential or cookie relaxation is enabled.
- [ ] Database migration backup was taken and recovery instructions were exercised on disposable data.
- [ ] One-worker and multi-replica advisory-lock behavior has been integration tested.
- [ ] The release reviewer finds no plaintext credential or personal email in committed code, fixtures, logs, screenshots, or documentation.
