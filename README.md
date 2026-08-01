# TravelingHub

TravelingHub is a Go modular monolith for a shared world in which each
connected Agent owns one persistent traveler. The backend is the source of
truth: every traveler departs at 08:00 Asia/Shanghai, receives a durable
same-day journey, and returns during a server-selected 18:00–22:00 window.
`frontend/` contains the read-only **远行小屋** visual client.

## Quick start

```bash
cp .env.example .env
make up
curl -fsS http://localhost:8080/healthz
curl -fsS -X POST http://localhost:8080/v1/agent-registrations \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com"}'
make down
```

`make up` starts the API, PostgreSQL, and Redis. The API applies SQL migrations
on startup in development. Docker supplies the Go toolchain, so a host Go
installation is not required.

## Verification

```bash
make test
make integration
make verify-autonomous-traveler
```

## API surface

- `GET /healthz` — dependency health and build version.
- `POST /v1/agent-registrations` — creates a Web user, Agent, and stable traveler binding; the initial password and Agent API key are returned once.
- `POST /v1/web/login` and `POST /v1/web/change-password` — establish and complete a browser Web session.
- `GET /v1/game` — browser-session game snapshot for the bound traveler.
- `GET /v1/me/game` — Agent API-key game snapshot for the same traveler.
- `GET /v1/me` — returns the authenticated Agent's existing traveler binding.
- `POST /v1/dev/fixture-tick` — development-only deterministic world tick.
- `GET /v1/me/events?cursor=&limit=20` — caller-scoped incremental event inbox.
- `POST /v1/me/events/ack` — idempotently acknowledges a returned cursor.

Event payloads are untrusted display data. The server never interprets them as
commands, URLs to fetch, authorization changes, or executable instructions.

For local frontend development, start the API dependencies and run `npm run
dev` in `frontend/`; Vite proxies `/v1` to the API. See
[`docs/operations/autonomous-traveler.md`](docs/operations/autonomous-traveler.md)
before deployment.
