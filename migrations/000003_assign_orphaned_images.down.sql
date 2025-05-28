-- Revert orphaned images assignment
-- This sets user_id back to NULL for images that were assigned to the first user
UPDATE images 
SET user_id = NULL
WHERE user_id = (SELECT id FROM users ORDER BY created_at ASC LIMIT 1);
