BEGIN;

DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP TRIGGER IF EXISTS update_files_updated_at ON files;

DROP TRIGGER IF EXISTS update_replicas_updated_at ON replicas;

DROP TYPE IF EXISTS user_role;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS files;

DROP TYPE IF EXISTS replica_status;

DROP TABLE IF EXISTS replicas;

DROP FUNCTION IF EXISTS update_updated_at_column();

COMMIT;