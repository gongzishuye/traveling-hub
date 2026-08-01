# Autonomous Traveler Rhythm Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make TravelingHub the authoritative, secure, autonomous daily-travel service for one traveler per registered Agent, while adapting the copied 远行小屋 frontend into a read-only web projection.
**Architecture:** Keep the Go service as a modular monolith. An identity module binds a Web user, Agent API key, and traveler; a journey module owns persistent daily state and reconciliation; one database-backed worker advances due transitions; HTTP exposes separate Agent-key and Web-cookie surfaces. React renders snapshots and never simulates outcome state.
**Tech Stack:** Go 1.25, net/http, PostgreSQL 16, Redis 7, Docker Compose, React 19, TypeScript, Vite, Vitest, Playwright.

---

## Outcome and non-negotiable rules

- A newly registered email creates exactly one Web user, Agent, and frog (the internal traveler record) in one database transaction.
- The first successful registration response is the sole place where the initial password and Agent API key are returned in plaintext. They must never be logged, stored in browser storage, or returned again.
- The browser authenticates only through a secure HttpOnly session cookie. Agent endpoints authenticate only through a high-entropy API key.
- Each eligible traveler has at most one daily_journeys row per Asia/Shanghai local date. At 08:00 it starts; its exact return time is selected once in [18:00, 22:00) and persisted.
- A periodic worker and snapshot reads both call the same reconciliation service. There are no per-traveler goroutines or browser timers deciding state.
- Departure, durable lifecycle event, album creation, and return transition are atomic. Retrying a worker, a request, or a process after a crash cannot duplicate a journey, event, or postcard.
- The copied visual assets and fixed template content stay local to frontend/; the backend selects and persists an identifier only. The frontend maps identifiers to existing text and assets and has no gameplay mutation controls.

## Target module and data shape

    Browser (Web session cookie) ─┐
                                  ├─ api/router.go
    Agent (Bearer API key) ───────┘       │
                                  ┌────────┴────────┐
                                  │ identity         │  users, sessions, agents
                                  │ journey          │  catalog, reconciliation, snapshot
                                  │ event + world    │  durable Agent event inbox
                                  └────────┬────────┘
                                           PostgreSQL
                                              ▲
                           one bounded worker │
                                              │
                                   journey.Service.Reconcile*

### Persistent model

The implementation will add the following bounded data model in migrations rather than adding a second service or event store:

| Record | Ownership and required constraints |
| --- | --- |
| users | normalized email is unique; Argon2id password hash; must_change_password; verification metadata |
| agents.user_id | one-to-one ownership for newly registered Agents; API-key digest is distinct from password hash |
| Redis session | opaque, random session ID stored only in a secure cookie; Redis value contains user ID and restricted/full state with TTL |
| daily_journeys | unique (frog_id, local_date); locked template/postcard/food IDs, departure/return timestamps, lifecycle status, next stage |
| events extensions | journey_id, journey_stage, and an idempotency key; existing cursor feed remains the Agent delivery surface |
| album_entries | unique journey_id, so one returned journey produces one postcard |

The existing world_ticks sequence will remain the cursor ordering source. Journey event writes will use transaction-aware world/event helpers so their mutation, tick, event, audience, and optional album entry commit together.

### Daily lifecycle

| Stage | Authoritative transition | Visible state |
| --- | --- | --- |
| Before the first eligible 08:00 | no current journey; retain prior returned journey for reading | home or prior returned |
| 08:00 local | insert locked daily journey and departure event | travelling |
| 25% and 50% of the locked trip duration | append the first and second template events | travelling |
| locked return time | append final template event, create album entry, mark returned | returned |
| Next day 08:00 | create the new daily journey; history remains in album | travelling |

The third existing template event is emitted at the return transition. This preserves the current three-event visual content without preserving the prototype's 15/30/45-second clock.

## Delivery sequence and gates

| Phase | Plan | Depends on | Exit gate |
| --- | --- | --- | --- |
| 1 | [Identity and credential boundary](./2026-08-01-autonomous-traveler-rhythm-phase-1-identity.md) | baseline | atomic registration, credentials split, forced first-password change, session security tests |
| 2 | [Journey domain and durable transitions](./2026-08-01-autonomous-traveler-rhythm-phase-2-journey-domain.md) | 1 | fixed-clock tests prove unique journeys, stable outcomes, atomic events and album entries |
| 3 | [Worker and Agent projection](./2026-08-01-autonomous-traveler-rhythm-phase-3-worker-agent.md) | 2 | offline worker drives state and Agent key sees only its own incremental events/snapshot |
| 4 | [Web API and deployment boundary](./2026-08-01-autonomous-traveler-rhythm-phase-4-web-api.md) | 1–3 | cookie-authenticated game endpoint resolves the owner and does not expose in-flight return time |
| 5 | [Read-only 远行小屋 adapter](./2026-08-01-autonomous-traveler-rhythm-phase-5-frontend.md) | 4 | frontend reads snapshots, retains visuals, and cannot mutate or locally simulate travel |
| 6 | [End-to-end UAT and release readiness](./2026-08-01-autonomous-traveler-rhythm-phase-6-uat-release.md) | 1–5 | every Agent, User, and System case from the PRD is evidenced |

## Cross-phase technical decisions

### Identity

- Normalize and validate email before database access. Repeated email registration returns a uniform accepted response without user, Agent, traveler, password, or key fields.
- Hash passwords with Argon2id and a per-password random salt. Store only the encoded parameterized hash; compare in constant time.
- Generate a 32-byte random Agent API key and opaque session IDs with crypto/rand; retain only SHA-256 digests for API keys and sessions.
- Keep legacy agents.token_digest compatibility through the migration, but stop the current Authenticate upsert behavior for new API calls. Authentication becomes lookup-only: no request may silently create an Agent/Frog.
- Cookie policy is HttpOnly; Secure; SameSite=Lax; Path=/. All state-changing cookie routes also verify same-origin Origin when present. Local browser development and Playwright run over HTTPS rather than downgrading the cookie flag.

### Reconciliation and concurrency

- Reuse world.Clock so tests inject a fixed instant and production uses UTC system time converted to Asia/Shanghai.
- Inject a deterministic random source in journey tests; production uses crypto/rand. Template and return time are persisted during the daily insert, never recomputed.
- The worker runs a short bounded loop. It first creates missing 08:00 journeys with INSERT ON CONFLICT, then claims due rows using FOR UPDATE SKIP LOCKED.
- A PostgreSQL advisory lock prevents duplicate whole-cycle scans in a multi-replica deployment. Database uniqueness and transaction idempotency remain the final correctness boundary if a lock is lost.
- Web and Agent snapshots reconcile only the caller's traveler before reading; they use the same row-locking code as the worker.

### Contract ownership

- GET /v1/game is cookie-only and derives user_id to agent_id to frog_id; it ignores any client-supplied IDs.
- GET /v1/me/game is Agent-key-only and returns the same snapshot shape for the authenticated Agent. Existing Agent event and acknowledgement APIs remain the incremental feed.
- A travelling snapshot omits return_at; a returned snapshot may include it as history.
- The frontend owns asset lookup and accessible display strings. Backend owns phase, dates, template/postcard IDs, already-visible event sequence, and album order.

## Program tasks

- [ ] Complete Phase 1 and record fresh unit/API evidence before changing journey scheduling.
- [ ] Complete Phase 2 and prove persistence/idempotency under a fixed Asia/Shanghai clock before starting a worker.
- [ ] Complete Phase 3 with an offline-worker test before exposing Web gameplay.
- [ ] Complete Phase 4 against a real cookie session before changing frontend/src/App.tsx.
- [ ] Complete Phase 5 while preserving visual baseline tests or deliberately update only behavior-dependent assertions.
- [ ] Complete Phase 6, attach the PRD's Agent/User/System evidence to docs/qa/005-autonomous-traveler-rhythm/, and update operational documentation.

## Definition of done

- [ ] All requirements FR-001 through FR-018 and acceptance criteria AC-001 through AC-018 in [the PRD](../../../specs/005-autonomous-traveler-rhythm.md) map to an automated test or documented manual security check.
- [ ] make test, make integration, frontend unit/lint/build, and Playwright E2E commands pass from a clean checkout.
- [ ] No plaintext initial password or Agent API key is present in source, test fixtures, logs, browser storage, non-first registration responses, or committed evidence.
- [ ] frontend/ contains source only and no nested Git repository, node_modules, or build output.
- [ ] README, API documentation, environment template, and deployment notes explain the new credentials, TLS/cookie constraint, worker ownership, migration order, backup, and rollback limitation.
