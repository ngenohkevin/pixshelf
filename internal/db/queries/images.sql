-- Users
-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    google_id, email, name, avatar_url
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET name = $2,
    avatar_url = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- Images
-- name: GetImage :one
SELECT * FROM images
WHERE id = $1 LIMIT 1;

-- name: ListImages :many
SELECT * FROM images
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountImages :one
SELECT COUNT(*) FROM images
WHERE user_id = $1;

-- name: SearchImages :many
SELECT * FROM images
WHERE user_id = $1 AND (name ILIKE $2 OR description ILIKE $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountSearchImages :one
SELECT COUNT(*) FROM images
WHERE user_id = $1 AND (name ILIKE $2 OR description ILIKE $2);

-- name: CreateImage :one
INSERT INTO images (
    name, description, file_path, mime_type, size_bytes, user_id
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateImage :one
UPDATE images
SET name = $2,
    description = $3,
    updated_at = NOW()
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1 AND user_id = $2;

-- name: GetImageByUser :one
SELECT * FROM images
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListImagesCursor :many
SELECT * FROM images
WHERE user_id = $1 AND id < $2
ORDER BY id DESC
LIMIT $3;

-- name: SearchImagesCursor :many
SELECT * FROM images
WHERE user_id = $1 AND id < $2 AND (name ILIKE $3 OR description ILIKE $3)
ORDER BY id DESC
LIMIT $4;
