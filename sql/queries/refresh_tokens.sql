-- name: CreateRefreshToken :one
insert into refresh_tokens (token, user_id, expires_at, created_at, updated_at, revoked_at)
VALUES ($1, $2, $3, NOW(), NOW(), $4)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
updated_at = NOW()
WHERE token = $1;