-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshTokenByToken :one
SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens
WHERE token = $1 LIMIT 1;

-- name: DeleteRefreshToken :execrows
DELETE FROM refresh_tokens
WHERE id = $1;

-- name: DeleteRefreshTokenByUserID :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;