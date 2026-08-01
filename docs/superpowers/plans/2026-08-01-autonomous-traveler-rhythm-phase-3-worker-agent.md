# Autonomous Traveler Rhythm — Phase 3: Worker and Agent Projection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Advance all travelers while no client is connected, and give an Agent API-key caller a safe, scoped snapshot plus its existing incremental event feed.
**Architecture:** A bounded journey worker owns one PostgreSQL advisory lock and invokes the Phase 2 reconciler on a configurable cadence. Agent HTTP routes use existing Bearer middleware and resolve only the authenticated Agent's frog.
**Tech Stack:** Go contexts and tickers, PostgreSQL advisory locks and row locks, net/http, Docker Compose integration tests.

---

## Files to add or change

- Add internal/journey/worker.go and internal/journey/worker_test.go.
- Modify internal/journey/service.go and repository.go for list-and-claim batch methods.
- Modify internal/platform/app/app.go, internal/platform/config/config.go, internal/platform/config/config_test.go, deploy/docker-compose.yml, and cmd/server/main.go.
- Add api/agent_game.go and api/agent_game_test.go; modify api/router.go.
- Add tests/integration/journey_worker_test.go and tests/integration/agent_game_test.go.
- Update README.md and docs/architecture.md.

## Worker contract

The process starts one worker with the app lifecycle. Every configured minute it attempts a dedicated PostgreSQL advisory lock, returns immediately if another replica owns it, and otherwise:

1. identifies frogs with no local-date journey once local time is at or after 08:00;
2. creates departures through the unique database key;
3. claims due stage rows in finite batches using FOR UPDATE SKIP LOCKED;
4. commits each transition with Phase 2 transactional reconciliation;
5. records only aggregate counts and error class, never payloads or credentials.

The worker shutdown waits for the current bounded cycle or context cancellation. A request path remains able to reconcile one traveler as recovery if a cycle was missed.

## Tasks

- [ ] **Write pure worker-cycle tests first.** In internal/journey/worker_test.go, supply a fake clock and repository seam to prove one cycle requests departures after 08:00, does nothing before 08:00, drains a bounded number of due rows, and returns when the advisory lock is unavailable. Run the focused test before implementation.

- [ ] **Implement the worker lifecycle.** Add worker.go with Start(context), RunOnce(context), and a small configurable cadence/batch size. Start a ticker only once; never create one ticker or goroutine per traveler. Re-run worker tests with race detection where supported.

- [ ] **Write real database concurrency tests.** In tests/integration/journey_worker_test.go, start two workers against the same Postgres and fixed 08:00 instant; assert one daily journey and one departure event. Repeat at return_at and assert one album entry/final event. Run the test and confirm it fails until real lock/claim code exists.

- [ ] **Implement database leadership and batch claiming.** Add repository methods that call pg_try_advisory_lock for the fixed worker key and release it reliably; select due rows with a deterministic order, a LIMIT, and FOR UPDATE SKIP LOCKED. Keep the daily unique index and stage deduplication as the final correctness guarantees. Re-run the integration test.

- [ ] **Write offline behavior test.** In tests/integration/journey_worker_test.go, create a registered traveler, make no Agent or Web calls, advance the controllable clock over 08:00 and return_at, invoke only the worker cycle, and assert AC-012. Include a delayed-cycle test that crosses all due stages.

- [ ] **Wire worker startup and graceful shutdown.** Extend App to own the journey service/worker; add worker cadence and batch settings to config plus tests of defaults and invalid values. Start it from the main lifecycle and stop before closing Redis/Postgres. Add Compose environment settings. Re-run make test and a Compose smoke test.

- [ ] **Write Agent game route tests first.** In api/agent_game_test.go, test GET /v1/me/game with a valid API key returns only that Agent's frog snapshot; missing/unknown key returns 401; query parameters cannot select another frog; travelling JSON lacks return_at. Run tests before routing.

- [ ] **Implement GET /v1/me/game.** Add api/agent_game.go, reuse the authenticated middleware, resolve principal Agent ID to frog, call ReconcileFrog before snapshot assembly, and serialize the same public schema planned for the Web endpoint. Do not add a mutation route. Re-run tests.

- [ ] **Prove incremental Agent feed compatibility.** Extend tests/integration/agent_game_test.go to create departures and returns through the worker, retrieve them via GET /v1/me/events, acknowledge the returned cursor, and prove a second fetch does not redeliver. Add a cross-Agent cursor attack assertion for AG-04 and AG-05.

- [ ] **Document operational behavior.** Update docs/architecture.md and README.md with cadence, Asia/Shanghai scope, advisory-lock behavior, bounded recovery semantics, and the Agent registration/key flow. Do not describe the worker as an exact wall-clock guarantee beyond its configured cadence.

- [ ] **Commit Phase 3 atomically.** Run make test and make integration after a clean Compose restart, inspect worker logs for secret redaction, then commit with a message such as feat: run autonomous journey worker.
