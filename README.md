# TravelingHub

TravelingHub is a Go modular-monolith foundation for a shared world in which
each connected Agent owns one persistent Frog. This repository deliberately
contains no travel gameplay, LLM narration, content feed, vector search, or
Agent-to-Agent direct messaging.

## Quick start

```bash
cp .env.example .env
make up
curl -fsS http://localhost:8080/healthz
curl -fsS -H 'Authorization: Bearer agent-alice' http://localhost:8080/v1/me
make down
```

`make up` starts the API, PostgreSQL, and Redis. The API applies SQL migrations
on startup in development. Docker supplies the Go toolchain, so a host Go
installation is not required.

## Verification

```bash
make test
make integration
```

## API surface

- `GET /healthz` — dependency health and build version.
- `GET /v1/me` — creates or returns the caller's stable Agent/Frog binding.
- `POST /v1/dev/fixture-tick` — development-only deterministic world tick.
- `GET /v1/me/events?cursor=&limit=20` — caller-scoped incremental event inbox.
- `POST /v1/me/events/ack` — idempotently acknowledges a returned cursor.

Event payloads are untrusted display data. The server never interprets them as
commands, URLs to fetch, authorization changes, or executable instructions.
