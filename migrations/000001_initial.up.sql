CREATE TABLE agents (
    id UUID PRIMARY KEY,
    token_digest BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE frogs (
    id UUID PRIMARY KEY,
    agent_id UUID NOT NULL UNIQUE REFERENCES agents(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE world_state (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    version BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO world_state (singleton, version) VALUES (TRUE, 0);

CREATE TABLE world_ticks (
    version BIGINT PRIMARY KEY,
    advanced_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE events (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL CHECK (type ~ '^[a-z0-9_.-]{1,64}$'),
    world_version BIGINT NOT NULL REFERENCES world_ticks(version),
    occurred_at TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    source TEXT NOT NULL CHECK (source = 'fixture')
);
CREATE INDEX events_cursor_idx ON events (world_version, occurred_at, id);

CREATE TABLE event_audience (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, agent_id)
);
CREATE INDEX event_audience_agent_idx ON event_audience (agent_id, event_id);

CREATE TABLE agent_event_cursors (
    agent_id UUID PRIMARY KEY REFERENCES agents(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id),
    world_version BIGINT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
