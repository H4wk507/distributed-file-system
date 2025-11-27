BEGIN;

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,
    size BIGINT NOT NULL,
    hash TEXT NOT NULL,
    content_type TEXT NOT NULL,
    owner_id UUID NOT NULL, -- TODO

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE replica_status AS ENUM ('synced', 'syncing', 'failed');

CREATE TABLE IF NOT EXISTS replicas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    file_id UUID NOT NULL,
    node_id UUID NOT NULL, -- TODO
    status replica_status NOT NULL DEFAULT 'syncing',

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE ON UPDATE CASCADE -- TODO
)

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_files_updated_at BEFORE UPDATE ON files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_replicas_updated_at BEFORE UPDATE ON replicas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;