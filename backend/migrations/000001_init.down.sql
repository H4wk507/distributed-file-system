DROP TABLE IF EXISTS files;

DROP TRIGGER IF EXISTS update_files_updated_at ON files;

DROP FUNCTION IF EXISTS update_updated_at_column();