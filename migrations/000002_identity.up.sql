CREATE TABLE users (
    id UUID PRIMARY KEY,
    email_normalized TEXT NOT NULL UNIQUE CHECK (email_normalized = lower(btrim(email_normalized)) AND length(email_normalized) BETWEEN 3 AND 320),
    password_hash TEXT NOT NULL,
    must_change_password BOOLEAN NOT NULL DEFAULT TRUE,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE agents
    ADD COLUMN user_id UUID UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN api_key_digest BYTEA;

UPDATE agents SET api_key_digest = token_digest WHERE api_key_digest IS NULL;
ALTER TABLE agents ALTER COLUMN api_key_digest SET NOT NULL;
CREATE UNIQUE INDEX agents_api_key_digest_idx ON agents(api_key_digest);
