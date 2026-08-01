# Autonomous Traveler Rhythm — Phase 1: Identity and Credential Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Atomically register one Web user, Agent, and traveler from an email, while separating browser-password/session authentication from Agent API-key authentication.
**Architecture:** Add an internal/identity module for credential generation, password hashing, sessions, and registration orchestration. Retain internal/agent and internal/frog as persistence-facing collaborators, but make Agent authentication lookup-only.
**Tech Stack:** Go 1.25, golang.org/x/crypto/argon2, PostgreSQL, Redis, net/http, Go httptest.

---

## Scope

This phase implements registration and login/password-change boundaries only. It does not create journeys, start a worker, or expose game data. That separation makes credential and ownership tests independent of simulation timing.

## Files to add or change

- Add migrations/000002_identity.up.sql and migrations/000002_identity.down.sql.
- Add internal/identity/model.go, password.go, password_test.go, credentials.go, credentials_test.go, session.go, session_test.go, service.go, and service_test.go.
- Modify internal/agent/service.go and add internal/agent/service_test.go.
- Modify internal/frog/service.go only to expose a transaction-safe insert helper used by registration.
- Modify internal/platform/app/app.go, internal/platform/config/config.go, internal/platform/config/config_test.go, and .env.example.
- Split router concerns by adding api/identity.go, api/web_auth.go, api/web_middleware.go, and api/identity_test.go; reduce api/router.go to route assembly and existing health/Agent routes.
- Add tests/integration/identity_test.go and update README.md with the non-secret API contract.

## Database contract

The identity migration creates users with id, email_normalized UNIQUE, password_hash, must_change_password, email_verified_at, timestamps; adds nullable unique agents.user_id referencing users; and adds an API-key digest column without dropping existing dev credentials. It uses a staged migration:

1. Add new columns and indexes without invalidating existing rows.
2. Backfill the API-key digest from the existing token_digest.
3. Leave legacy rows unlinked rather than inventing owners; only registered accounts are Web-login eligible.

The down migration reverses only structures introduced here when no later migration depends on them.

## Tasks

- [ ] **Write migration contract tests.** In internal/platform/postgres/migrate_test.go, add coverage that applies 000001 then 000002 to an empty PostgreSQL database and verifies users, Agent ownership uniqueness, and API-key digest indexes. Run the focused test and confirm it fails before the migration exists.

- [ ] **Implement the reversible identity migration.** Add the migration pair with the staged schema above, including checks for normalized non-empty email and password-change state. Re-run the focused migration test until it passes; then run make integration.

- [ ] **Write password behavior tests first.** In internal/identity/password_test.go, test Argon2id hashing/verification, rejection of blank or weak replacement passwords, different salts for equal input, malformed hash rejection, and constant-time verification path. Run go test ./internal/identity -run Password and observe failure.

- [ ] **Implement password and credential primitives.** In password.go encode Argon2id parameters, salt, and digest in one versioned value; in credentials.go generate opaque initial passwords, 32-byte API keys, and session IDs with crypto/rand. Digest API keys/session IDs with SHA-256. Keep plaintext values only in result structs, never persisted fields. Re-run focused tests and add credential entropy/format tests.

- [ ] **Write Agent lookup-only authentication tests.** In internal/agent/service_test.go, prove a valid stored API-key digest resolves its Agent, unknown and blank keys fail, and no agents row is inserted on failed or unknown authentication. Run the test before changing the service.

- [ ] **Replace request-driven Agent/Frog creation.** Modify internal/agent/service.go to authenticate by digest lookup and return Principal only. Change GET /v1/me to load the registered Agent's existing frog and return a generic not-found error for a legacy unbound Agent; it must not create either record. Update its tests and run go test ./internal/agent ./internal/frog.

- [ ] **Write registration transaction tests.** In internal/identity/service_test.go, test a new normalized email returns a user/Agent/frog binding plus plaintext credentials; test an injected failure after user creation rolls back all rows; test a duplicate email creates nothing and yields a credential-free duplicate result. Run the test and confirm it fails.

- [ ] **Implement transactional registration.** In internal/identity/service.go, validate/normalize email, hash the temporary password, generate API key, then create users, agents, and frogs in one sql.Tx. Add a narrow frog.InsertForAgentTx helper rather than opening nested transactions. Resolve a unique-email conflict to a deliberately non-enumerating duplicate result. Re-run focused tests and add a concurrent same-email test using two goroutines and database assertions.

- [ ] **Write Web session and first-login tests.** In internal/identity/session_test.go and service_test.go, cover opaque Redis session creation, expiry, user lookup, restricted first-login state, password change invalidating the restricted session, and old-password rejection. Run tests before the implementation.

- [ ] **Implement session and login service.** Store only session digests and user ID/session state in Redis with configurable TTL. Authenticate email/password, issue restricted or full sessions according to must_change_password, and atomically replace the password hash and clear that flag on change. Re-run identity tests.

- [ ] **Write verification-policy tests.** Add tests for the explicit development bypass, pending unverified production user, verified production user, expired verification token, and production startup without a configured verification delivery adapter. Run them before adding the policy.

- [ ] **Implement email verification boundary.** Add a random, hashed, expiring verification token store and a narrow mail-delivery interface. Development may bypass verification only with an explicit development-only config flag; production must fail closed until the user is verified and must refuse to start if its delivery adapter is absent. Keep the provider-specific adapter outside the domain service and record the provider selection as a Phase 6 deployment prerequisite. Re-run identity tests.

- [ ] **Write HTTP contract tests first.** In api/identity_test.go, assert malformed emails are 400; first registration returns 201 and precisely the allowed credential fields; duplicate registration returns a generic accepted response with no IDs/credentials; unverified production users cannot start a Web session; login writes a secure HttpOnly SameSite cookie; restricted sessions reject game-space routes and permit only password change. Assert no response body includes a credential outside the first successful registration.

- [ ] **Implement HTTP handlers and middleware.** Add POST /v1/agent-registrations, POST /v1/web/login, POST /v1/web/change-password, and cookie middleware. Set Secure, HttpOnly, SameSite=Lax, and Path=/ unconditionally; reject unsafe cross-origin write requests. Wire identity/session services in internal/platform/app/app.go; add explicit session TTL and cookie-name config keys with secure defaults. Re-run API unit tests.

- [ ] **Prove Phase 1 in containers.** Add tests/integration/identity_test.go for AG-01, AG-02, AG-03, US-01, SY-03, and SY-04 portions of the PRD. Run make test, make integration, and a manual curl flow that redacts returned credentials. Update README.md and .env.example with endpoint names and configuration names, never example secrets.

- [ ] **Commit Phase 1 atomically.** Review git diff --check, verify no credential literals are staged, then commit the migration, code, tests, and docs together with a message such as feat: add traveler identity boundary.
