-- migrations/000004_optimize_indexes.down.sql
DROP INDEX IF EXISTS idx_images_user_created;
DROP INDEX IF EXISTS idx_images_name_trgm;
DROP INDEX IF EXISTS idx_images_updated;