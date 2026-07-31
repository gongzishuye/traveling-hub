# TravelingHub Foundation Architecture

TravelingHub is a single Go process with explicit domain modules. PostgreSQL is
the source of truth; Redis is provisioned as the future queue/cache boundary.

`agent` authenticates a development bearer token by storing only its SHA-256
digest. `frog` atomically binds one Frog to one Agent. `world` advances a
monotonic version. `event` stores immutable, audience-scoped events and an
Agent cursor. `simulation` is fixture-only and may create one deterministic
event after a world tick.

The public API does not reach into SQL directly. It authenticates a principal,
then calls the relevant domain service. This preserves a modular-monolith
boundary that can later be split without changing the public contract.
