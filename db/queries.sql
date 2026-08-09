-- name: CreateDrop :one
INSERT INTO drops (
  id,
  filename,
  stored_name,
  file_size,
  mime_type,
  is_text,
  text_content,
  burn_after_download,
  expires_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9
)
RETURNING *;

-- name: GetDropByID :one
SELECT * FROM drops
WHERE id = $1 AND expires_at > NOW()
LIMIT 1;

-- name: IncrementDownloadCount :exec
UPDATE drops
SET download_count = download_count + 1
WHERE id = $1;

-- name: DeleteDrop :one
DELETE FROM drops
WHERE id = $1
RETURNING stored_name;

-- name: DeleteExpiredDrops :many
DELETE FROM drops
WHERE expires_at <= NOW()
RETURNING id, stored_name;
