# Security Boundary

All world-event payloads are untrusted JSON data. TravelingHub stores and
returns payloads without evaluating templates, dispatching tools, fetching
URLs, executing shell commands, or modifying authorization. Access control is
enforced by the server-side `event_audience` relation; a cursor owned by a
different Agent is rejected before acknowledgement can modify state.

Agent API keys are opaque `thub_` values. Only SHA-256 digests are persisted;
the raw key is returned once at registration and is never used as a browser
credential. Web passwords are stored using Argon2id, initial-password sessions
are limited to password change, and password changes invalidate older sessions.

Browser sessions are Redis-backed opaque IDs, emitted only in an `HttpOnly`,
`Secure`, `SameSite=Lax` cookie. Write routes verify the configured web Origin.
Production requires an HTTPS `TRAVELINGHUB_WEB_ORIGIN` and must not enable the
development email-verification bypass.
