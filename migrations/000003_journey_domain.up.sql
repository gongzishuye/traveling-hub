ALTER TABLE events DROP CONSTRAINT events_source_check;
ALTER TABLE events ADD CONSTRAINT events_source_check CHECK (source IN ('fixture', 'journey'));
ALTER TABLE events ADD COLUMN journey_id UUID;
ALTER TABLE events ADD COLUMN journey_stage SMALLINT;
ALTER TABLE events ADD COLUMN deduplication_key TEXT UNIQUE;
CREATE INDEX events_journey_stage_idx ON events (journey_id, journey_stage);

CREATE TABLE daily_journeys (
    id UUID PRIMARY KEY,
    frog_id UUID NOT NULL REFERENCES frogs(id) ON DELETE CASCADE,
    local_date DATE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('travelling', 'returned')),
    template_id TEXT NOT NULL,
    postcard_id TEXT NOT NULL,
    food_id TEXT NOT NULL,
    departed_at TIMESTAMPTZ NOT NULL,
    return_at TIMESTAMPTZ NOT NULL CHECK (return_at > departed_at),
    next_stage SMALLINT NOT NULL CHECK (next_stage BETWEEN 1 AND 3),
    returned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (frog_id, local_date)
);
CREATE INDEX daily_journeys_due_idx ON daily_journeys (status, return_at);

CREATE TABLE album_entries (
    id UUID PRIMARY KEY,
    frog_id UUID NOT NULL REFERENCES frogs(id) ON DELETE CASCADE,
    journey_id UUID NOT NULL UNIQUE REFERENCES daily_journeys(id) ON DELETE CASCADE,
    postcard_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX album_entries_frog_created_idx ON album_entries (frog_id, created_at, id);
