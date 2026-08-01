# Autonomous traveler operations

## Required configuration

- PostgreSQL is the durable source of truth. Back it up before applying a new
  migration. Restoring that backup is the recovery path for a failed release;
  do not run journey down-migrations against user-created data.
- Redis stores browser sessions. A Redis loss logs users out but must not
  change a traveler, journey, event, or album entry.
- Set `TRAVELINGHUB_WEB_ORIGIN` to the exact HTTPS public origin in production.
  The reverse proxy must terminate HTTPS before browser traffic reaches the
  Go service, because Web sessions are `Secure`, `HttpOnly`, and `SameSite=Lax`.
- Set `TRAVELINGHUB_AUTO_VERIFY_EMAIL=false` in production. Do not expose a
  public deployment until an email-verification delivery adapter is configured
  and has passed a disposable-address staging test.
- `TRAVELINGHUB_JOURNEY_WORKER_INTERVAL` controls the bounded reconciliation
  cadence (default one minute); `TRAVELINGHUB_JOURNEY_WORKER_BATCH` bounds each
  database pass (default 100 travelers).

## Start, migrate, and restart

1. Take and verify a PostgreSQL backup.
2. Deploy the application with `TRAVELINGHUB_AUTO_MIGRATE=true` for the
   migration release, or run the migration command once in a controlled job.
3. Check `GET /healthz` only after PostgreSQL and Redis are healthy.
4. Restarting one or many replicas is safe: the Worker uses a PostgreSQL
   advisory lock, and the unique `(frog_id, local_date)` constraint plus event
   deduplication keys make repeated cycles idempotent.
5. After a successful migration, disable automatic migrations for normal
   steady-state deployments if your release policy requires separate jobs.

## Observability and triage

| Symptom | Check first | Safe response |
| --- | --- | --- |
| A traveler did not depart | Worker process running, advisory-lock holder, database connectivity, local `Asia/Shanghai` time | Restore worker/database connectivity; the next cycle or snapshot read reconciles the missed departure. |
| Duplicate journey or postcard suspected | `daily_journeys` uniqueness, `album_entries(journey_id)`, lifecycle-event deduplication key | Do not edit rows manually; inspect duplicate writes and retain the first durable record. |
| Browser cannot stay logged in | HTTPS termination, exact web origin, Redis availability, cookie domain/path | Fix proxy/origin/Redis, then ask user to sign in again. |
| Frontend cannot reach API | Same-origin proxy or deployed API origin, `GET /healthz`, browser network response | Restore API route; do not make the frontend generate local travel state. |

Never put an email address, API key, password, session ID, or event payload in
routine logs or alerts. Monitor aggregate worker-cycle duration, candidate
counts, transition counts, lock contention, reconciliation failures, and
authentication failure classes instead.

## Time-zone constraint

The first release uses `Asia/Shanghai` for every traveler. A future per-traveler
time zone is a product migration, not a deployment-time environment override.
