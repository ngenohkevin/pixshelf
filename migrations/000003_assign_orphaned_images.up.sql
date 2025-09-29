-- Assign orphaned images (those with NULL user_id) to the first user
-- This is needed when adding authentication to an existing system with images
UPDATE images 
SET user_id = (SELECT id FROM users ORDER BY created_at ASC LIMIT 1)
WHERE user_id IS NULL 
AND EXISTS (SELECT 1 FROM users);
