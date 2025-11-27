BEGIN;

DROP TRIGGER IF EXISTS update_files_updated_at ON files;

DROP TRIGGER IF EXISTS update_replicas_updated_at ON replicas;

DROP TABLE IF EXISTS files;

DROP TYPE IF EXISTS replica_status;

DROP TABLE IF EXISTS replicas;

DROP FUNCTION IF EXISTS update_updated_at_column();

COMMIT;