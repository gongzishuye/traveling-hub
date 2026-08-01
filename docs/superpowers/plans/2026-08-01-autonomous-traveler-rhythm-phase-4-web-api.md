# Autonomous Traveler Rhythm — Phase 4: Web API and Deployment Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Serve a cookie-authenticated, read-only game snapshot that always resolves the authenticated Web user to its own traveler and safely recovers missed transitions before responding.
**Architecture:** Build the Web game handler on Phase 1 cookie middleware and the Phase 2 journey service. The handler never accepts owner identifiers from the client, and it returns the same versioned snapshot DTO used by the Agent route.
**Tech Stack:** Go 1.25, net/http, PostgreSQL/Redis integration tests, HTTPS cookie deployment configuration.

---

## API contract frozen by this phase

GET /v1/game requires a full Web session and returns:

    {
      "frog_id": "uuid",
      "server_time": "RFC3339 timestamp with +08:00 offset",
      "local_date": "YYYY-MM-DD",
      "phase": "home | travelling | returned",
      "journey": {
        "template_id": "catalog ID",
        "departed_at": "timestamp",
        "return_at": "timestamp only after returned"
      },
      "events": [{"stage": 0, "text": "旅人出发了。"}],
      "album_postcard_ids": ["postcard ID"]
    }

No request parameter selects user, Agent, frog, phase, template, or return time. A restricted initial-password session gets a stable authorization error and must change its password before this route succeeds.

## Files to add or change

- Add api/web_game.go and api/web_game_test.go.
- Modify api/router.go and api/web_middleware.go from Phase 1.
- Modify internal/identity/model.go and service.go only if a strongly typed full-session principal is still missing.
- Modify internal/journey/snapshot.go and snapshot_test.go only to finalize JSON field names.
- Add tests/integration/web_game_test.go.
- Update .env.example, README.md, docs/security.md, and docs/architecture.md.

## Tasks

- [ ] **Write snapshot serialization tests.** Extend internal/journey/snapshot_test.go to marshal home, travelling, and returned snapshots. Assert server_time is Shanghai-local RFC3339, the travelling form lacks the return_at property entirely, and event/album ordering is stable. Run the focused test before finalizing tags.

- [ ] **Finalize one public snapshot DTO.** Put JSON tags and optional fields in internal/journey/snapshot.go, then have both Web and Agent handlers serialize that type. Do not duplicate anonymous map response shapes. Re-run unit tests.

- [ ] **Write Web ownership route tests first.** In api/web_game_test.go, create two registered users with full sessions and journeys. Assert each GET /v1/game returns only its own frog. Add agent_id and frog_id query parameters to the request and assert the result is unchanged. Test absent, expired, and restricted cookies separately. Run before handler implementation.

- [ ] **Implement full-session middleware and route.** Add a Web handler that reads the session principal, rejects restricted sessions, resolves user_id through the one-to-one Agent binding, finds that Agent's frog, calls ReconcileFrog, and returns the shared snapshot DTO. Do not accept an ID from URL, header, body, or cookie payload. Re-run route tests.

- [ ] **Write read-before-return recovery tests.** In api/web_game_test.go, create an overdue travelling journey without running the worker, then call GET /v1/game. Assert it returns returned state with one postcard and subsequent requests do not add records. Add a similar 08:00 missed-departure test. Run before relying on service wiring.

- [ ] **Wire reconciliation into the read path.** Ensure Web and Agent routes call the same narrow reconciliation method before their snapshot query. Propagate database errors as a generic 500 response without leaking query details or credentials. Re-run unit/API tests.

- [ ] **Write cookie and origin negative tests.** Cover Secure, HttpOnly, SameSite=Lax, Path, expiry/logout behavior if logout is present, and mismatched Origin rejection on password change. Cover cookie routes with an Agent API key and Agent routes with a Web cookie to prove credential surfaces cannot be exchanged.

- [ ] **Implement production deployment settings.** In config and .env.example, require externally terminated HTTPS or direct TLS before serving browser login/game traffic. Document a same-origin reverse-proxy deployment and an HTTPS Vite development setup. The service must not add Access-Control-Allow-Credentials for arbitrary origins.

- [ ] **Add integration test evidence.** In tests/integration/web_game_test.go, cover AC-005, AC-009, AC-011, AC-014, AC-015, and SY-03 with real Postgres/Redis. Use generated runtime credentials and redact them from failures/logs.

- [ ] **Update security documentation.** Add exact credential boundary, session expiry/invalidation, TLS requirement, cookie attributes, origin checks, and return-time privacy rules to docs/security.md. Update API examples in README.md with schema placeholders, never working secrets.

- [ ] **Commit Phase 4 atomically.** Run make test and make integration after a clean Compose start. Inspect response headers and log output, then commit with a message such as feat: expose secure web game snapshots.
