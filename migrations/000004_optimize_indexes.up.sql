-- migrations/000004_optimize_indexes.up.sql
CREATE INDEX CONCURRENTLY idx_images_user_created ON images(user_id, created_at DESC);
CREATE INDEX CONCURRENTLY idx_images_name_trgm ON images USING gin(name gin_trgm_ops);
CREATE INDEX CONCURRENTLY idx_images_updated ON images(updated_at DESC);

-- Enable pg_trgm for better search
CREATE EXTENSION IF NOT EXISTS pg_trgm;