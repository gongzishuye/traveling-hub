# Foundation Acceptance Evidence

Run these commands after `make up`:

```bash
curl -fsS http://localhost:8080/healthz
curl -fsS -H 'Authorization: Bearer agent-alice' http://localhost:8080/v1/me
make test
make integration
```

The integration suite proves stable Agent/Frog binding, deterministic event
creation, incremental cursor acknowledgement, cross-Agent isolation, and the
untrusted-event data boundary.

## Verified locally — 2026-07-31

- `go test ./...` passed in the Go 1.25 container.
- `go test -tags=integration ./tests/integration -count=1` passed against
  PostgreSQL 16 and Redis 7 containers.
- `GET /healthz` returned `{"status":"ok","version":"dev"}`.
- `docker build --pull=false -t travelinghub-app .` produced the runnable
  binary image; its `/healthz`, identity, fixture-tick, and inbox endpoints
  were exercised against the same PostgreSQL and Redis containers.
- Repeating `GET /v1/me` with the same bearer token returned the same Agent ID
  and Frog ID; a fixture tick produced an event; acknowledging its cursor
  returned HTTP 204 and the subsequent inbox was empty.
