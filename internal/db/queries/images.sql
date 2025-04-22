-- name: GetImage :one
SELECT * FROM images
WHERE id = $1 LIMIT 1;

-- name: ListImages :many
SELECT * FROM images
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountImages :one
SELECT COUNT(*) FROM images;

-- name: SearchImages :many
SELECT * FROM images
WHERE name ILIKE $1 OR description ILIKE $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSearchImages :one
SELECT COUNT(*) FROM images
WHERE name ILIKE $1 OR description ILIKE $1;

-- name: CreateImage :one
INSERT INTO images (
    name, description, file_path, mime_type, size_bytes
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateImage :one
UPDATE images
SET name = $2,
    description = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1;
