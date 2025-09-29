-- Remove user_id column from images table
ALTER TABLE images DROP COLUMN IF EXISTS user_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_images_user_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_google_id;

-- Drop users table
DROP TABLE IF EXISTS users;
