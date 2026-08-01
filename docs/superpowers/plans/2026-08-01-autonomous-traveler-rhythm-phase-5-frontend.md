# Autonomous Traveler Rhythm — Phase 5: Read-only 远行小屋 Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Preserve the current 远行小屋 visual experience while replacing all local gameplay state, random selection, and timing with a secure Web snapshot from TravelingHub.
**Architecture:** Keep frontend as an independently runnable Vite/React application. Add a narrow HTTP client and snapshot-to-view-model adapter; App handles login, forced password change, loading, errors, and controlled polling. Existing visual components receive read-only props.
**Tech Stack:** React 19, TypeScript, Vite, Vitest, Testing Library, Playwright.

---

## Frontend behavior contract

- Initial load requests GET /v1/game with credentials included. A missing session presents login; a restricted session presents change-password; a full session renders the game.
- While travelling, refresh every 60 seconds. While home or returned, refresh every five minutes and provide an accessible manual retry action after an error. Polling only retrieves a snapshot; it never advances state.
- A snapshot maps template_id and postcard IDs to the unchanged local catalog. Unknown IDs show a content-unavailable error and retain the last valid rendered snapshot; they do not choose a replacement.
- JourneyTimeline receives server-visible events in stage order. AlbumStrip receives server album cards. GardenStage receives the server phase.
- TravelJournal becomes a read-only daily note. Food radio controls, departure, restart, useJourneyTimer, local RNG, and local transition reducer calls are removed from production code.

## Files to add or change

- Add frontend/src/api/client.ts, frontend/src/api/types.ts, frontend/src/api/client.test.ts, frontend/src/domain/snapshot.ts, frontend/src/domain/snapshot.test.ts, frontend/src/hooks/useGameSnapshot.ts, and frontend/src/hooks/useGameSnapshot.test.ts.
- Add frontend/src/components/LoginScreen.tsx, LoginScreen.test.tsx, ChangePasswordScreen.tsx, ChangePasswordScreen.test.tsx, and GameError.tsx.
- Modify frontend/src/App.tsx, App.test.tsx, domain/travel.ts, domain/travel.test.ts, domain/journeyCatalog.ts, domain/journeyCatalog.test.ts, components/TravelJournal.tsx, JourneyTimeline.tsx, PostcardReveal.tsx, AlbumStrip.tsx where prop changes require it, and styles.css.
- Delete frontend/src/hooks/useJourneyTimer.ts and frontend/src/hooks/useJourneyTimer.test.ts after replacement tests pass.
- Modify frontend/vite.config.ts, frontend/playwright.config.ts, frontend/e2e/journey.spec.ts, frontend/e2e/visual-baseline.spec.ts, frontend/package.json, frontend/README.md, and root Makefile.

## Tasks

- [ ] **Write API client tests first.** In frontend/src/api/client.test.ts, mock fetch and assert credentials are included, only expected endpoints are called, 401 maps to unauthenticated, restricted-session authorization maps to password-change required, network failures remain errors, and no Agent API key is accepted or stored. Run npm test from frontend and confirm failure.

- [ ] **Implement the narrow Web client.** In api/client.ts, implement login, change-password, and getGameSnapshot with typed request/response parsing. Use relative paths only; never accept, persist, or emit Agent credentials. Re-run client tests.

- [ ] **Write catalog mapping tests before adapter code.** In domain/snapshot.test.ts, use real IDs from journeyCatalog.ts and assert a returned backend template maps to its existing three event texts and postcard asset. Assert unknown template/postcard IDs are typed failures. Include a parity assertion for all 18 template IDs.

- [ ] **Implement snapshot-to-view-model mapping.** Add domain/snapshot.ts; define backend snapshot types separately from presentational cards/events so no component depends on backend date math. Remove travel.ts functions that mutate or advance a game; retain only shared display types if still useful. Re-run mapping tests.

- [ ] **Write polling hook tests first.** In hooks/useGameSnapshot.test.ts, use fake timers to assert initial fetch, 60-second travelling refresh, five-minute non-travelling refresh, no overlapping requests, cleanup on unmount, preservation of last valid snapshot on error, and no setInterval-based journey advancement. Run tests before hook implementation.

- [ ] **Implement read-only snapshot hook.** Add useGameSnapshot.ts with cancellation via AbortController, phase-aware polling, retry state, and no local state transition logic. Re-run hook tests.

- [ ] **Write login and forced-password UI tests.** In component tests, assert the password field is not prefilled or retained after failure, login submits email/password only, restricted state exposes only change-password actions, and success transitions to snapshot loading. Test accessible labels, focus management, and error messages.

- [ ] **Implement auth screens.** Add LoginScreen and ChangePasswordScreen with explicit screen state in App. Do not render game panels before a full session exists. Do not place secrets in URLs, localStorage, sessionStorage, console output, or test IDs.

- [ ] **Write read-only component tests.** Update TravelJournal and App tests to assert no selectable food, departure, or restart control exists; returned postcard opening remains an optional view interaction only. Assert GardenStage, JourneyTimeline, and AlbumStrip receive their values from a supplied snapshot, not elapsed seconds.

- [ ] **Refactor App into snapshot projection.** Replace createInitialGame, startJourney, advanceJourney, startNextJourney, and useJourneyTimer production use in App.tsx. Pass mapped snapshot props into existing visual components, retaining art, layout, panel behavior, album viewing, and accessibility. Remove the timer hook files only after tests pass.

- [ ] **Add development transport.** Configure vite.config.ts with an explicit API proxy target and HTTPS local server configuration. Add documented environment variables for proxy target and certificate paths; keep certificates, node_modules, and dist ignored. Do not configure permissive credentialed CORS.

- [ ] **Replace prototype E2E journey test.** Rewrite frontend/e2e/journey.spec.ts to drive a controlled API fixture through login, forced password change, home, travelling, returned, refresh, cross-browser reload, expired session, API error, and unknown template flows. Assert no local timer changes phase and no return time appears while travelling.

- [ ] **Rebaseline only intentional visuals.** Run existing visual-baseline.spec.ts against deterministic snapshots. Retain scene and asset coverage; update snapshots only for removed controls and new auth/error states, reviewing every image diff.

- [ ] **Add root verification commands.** Make frontend install/test/lint/build/E2E commands callable from the repository root without committing build output. Run npm ci, npm test, npm run lint, npm run build, and Playwright from frontend plus make test and make integration.

- [ ] **Commit Phase 5 atomically.** Verify frontend/ has no node_modules, dist, test-results, or nested .git directory; inspect git diff --check; commit with a message such as feat: render autonomous traveler snapshots.
