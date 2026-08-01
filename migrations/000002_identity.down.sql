DROP INDEX IF EXISTS agents_api_key_digest_idx;
ALTER TABLE agents DROP COLUMN IF EXISTS api_key_digest;
ALTER TABLE agents DROP COLUMN IF EXISTS user_id;
DROP TABLE IF EXISTS users;
