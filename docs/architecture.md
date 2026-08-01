# TravelingHub Foundation Architecture

TravelingHub is a single Go process with explicit domain modules. PostgreSQL is
the source of truth; Redis is provisioned as the future queue/cache boundary.

`identity` atomically creates one Web user, Agent, and traveler and keeps Web
sessions separate from Agent API keys. `agent` authenticates an Agent API key
by its SHA-256 digest. `frog` binds one Frog to one Agent. `journey` owns the
durable `Asia/Shanghai` daily rhythm and its bounded, advisory-lock-protected
Worker. `world` advances a monotonic version. `event` stores immutable,
audience-scoped events and an Agent cursor. `simulation` is fixture-only and
may create one deterministic event after a world tick.

The Web and Agent game endpoints call the same journey snapshot service. That
service reconciles missed transitions before projecting a read-only view, so a
browser, Agent, or Worker can be offline without moving authority to the
client.

The public API does not reach into SQL directly. It authenticates a principal,
then calls the relevant domain service. This preserves a modular-monolith
boundary that can later be split without changing the public contract.
