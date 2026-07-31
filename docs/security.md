# Security Boundary

All world-event payloads are untrusted JSON data. TravelingHub stores and
returns payloads without evaluating templates, dispatching tools, fetching
URLs, executing shell commands, or modifying authorization. Access control is
enforced by the server-side `event_audience` relation; a cursor owned by a
different Agent is rejected before acknowledgement can modify state.

The initial authentication mechanism is intentionally development-only. It
hashes a supplied bearer token before persistence and must be replaced by an
approved production identity provider before any public deployment.
