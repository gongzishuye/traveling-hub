# Autonomous Traveler Rhythm — Phase 2: Journey Domain and Durable Transitions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist deterministic daily journey state, fixed-template events, and postcard history with exactly-once transition semantics under an injected Asia/Shanghai clock.
**Architecture:** Add an internal/journey domain module with a server-owned template catalog, repository, reconciler, and snapshot mapper. Refactor the existing world/event services minimally to append journey events inside the same SQL transaction as journey and album updates.
**Tech Stack:** Go 1.25, PostgreSQL transactions, UUIDs, existing world.Clock, existing event cursor model.

---

## Scope and decisions

- The server owns an immutable catalog whose identifiers exactly match the existing frontend journeyCatalog.ts IDs and postcard IDs. The first version includes the current 18 templates; selection is server-side.
- A returned journey remains the displayed journey before the next day's departure. A missing current-day journey before 08:00 means home only when no prior returned journey exists.
- For a trip from 08:00 to return_at, template content stages occur at 25%, 50%, and 100% of duration. Stage zero is the departure message.
- Production selection uses cryptographic randomness. Tests inject a deterministic selector. Both template and exact return instant are stored before any response.

## Files to add or change

- Add migrations/000003_journey_domain.up.sql and migrations/000003_journey_domain.down.sql.
- Add internal/journey/model.go, catalog.go, catalog_test.go, repository.go, service.go, service_test.go, snapshot.go, and snapshot_test.go.
- Modify internal/world/service.go and add transaction-helper tests in internal/world/service_test.go.
- Modify internal/event/model.go, internal/event/service.go, and internal/event/model_test.go.
- Add tests/integration/journey_domain_test.go.

## Database contract

The migration creates:

- daily_journeys(id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage, returned_at, created_at, updated_at) with status and stage checks and unique (frog_id, local_date).
- album_entries(id, frog_id, journey_id UNIQUE, postcard_id, created_at), ordered by created_at then id.
- nullable events.journey_id, nullable events.journey_stage, and nullable unique events.deduplication_key; an index supports journey_id and journey_stage lookups.
- an expanded event source constraint allowing fixture and journey.

No second event store is introduced: the existing Agent cursor feed remains the single durable event delivery record.

## Tasks

- [ ] **Write schema migration tests.** Extend internal/platform/postgres/migrate_test.go to assert the daily uniqueness key, one-album-per-journey key, journey-event index, and expanded event-source constraint after applying 000003. Run the focused test and confirm failure.

- [ ] **Implement the journey migration.** Add the up/down migration with named constraints and indexes. Preserve all existing event rows and world ticks. Run the migration test and make integration.

- [ ] **Write catalog parity tests.** In internal/journey/catalog_test.go, assert every server template ID, postcard ID, food ID, and three event texts matches a checked-in expected list copied from the frontend catalog. Test unique IDs and valid postcard IDs. Run it before adding the catalog.

- [ ] **Implement the server catalog.** Add internal/journey/catalog.go and model.go with the 18 fixed templates, no network/content dependency, and lookup functions that reject unknown IDs. Re-run catalog tests. Phase 5 adds a matching TypeScript parity test.

- [ ] **Write time-boundary unit tests first.** In internal/journey/service_test.go, with a fixed world.Clock and deterministic selector, cover 07:59, exactly 08:00, 08:01, 25%, 50%, return_at, 23:59, next-day 07:00, and next-day 08:01. Assert local date is always calculated in Asia/Shanghai even when the source instant is UTC.

- [ ] **Implement journey construction and snapshot model.** service.go calculates local dates, chooses a template and return instant in [18:00, 22:00), derives stage due times, and maps repository data to a DTO with phase, server time, visible events, and album postcard IDs. snapshot.go omits return_at for travelling and rejects unknown catalog IDs. Re-run unit tests.

- [ ] **Write transaction failure and idempotency tests.** Add tests that call reconciliation repeatedly, concurrently, and with an injected failure just before commit. Assert one journey, locked fields unchanged, and no partial journey/event/album rows. Run tests before transaction helpers exist.

- [ ] **Make world/event append transaction-aware.** Refactor internal/world/service.go so Advance delegates to an AdvanceTx(ctx, tx, at) helper. Add event.Service.AppendTx that writes events/audience through the caller transaction and supports a deduplication key. Keep public Append as a compatibility wrapper for the fixture endpoint. Update source validation and cursor tests.

- [ ] **Implement repository locking and reconciliation.** In internal/journey/repository.go, insert departures with ON CONFLICT (frog_id, local_date) DO NOTHING, lock a journey row before applying a due stage, and write the world tick, event/audience payload, and album entry in that one transaction. Use journey_id:stage as the deduplication key. Re-run idempotency tests.

- [ ] **Write snapshot isolation tests.** In internal/journey/snapshot_test.go, create journeys for two frogs and prove a service call for one never returns the other’s event or album row; prove return time is absent during travel and present only after return.

- [ ] **Implement scoped read methods.** Add repository methods that load only by authoritative frog ID, query visible journey events by events.journey_id, and order album entries stably. Re-run snapshot tests.

- [ ] **Run domain integration evidence.** Add tests/integration/journey_domain_test.go for AC-001, AC-002, AC-003, AC-014, and SY-02 using a fixed clock and real Postgres. Run make test and make integration; record no browser or worker claim yet.

- [ ] **Commit Phase 2 atomically.** Review SQL down migration and git diff --check; commit the migration, transaction refactor, domain tests, and catalog together with a message such as feat: persist autonomous daily journeys.
