BEGIN;

CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'user',

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL, -- TODO: unique?
    size BIGINT NOT NULL,
    hash TEXT NOT NULL,
    content_type TEXT NOT NULL,
    owner_id UUID NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
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

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_files_updated_at BEFORE UPDATE ON files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_replicas_updated_at BEFORE UPDATE ON replicas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


COMMIT;