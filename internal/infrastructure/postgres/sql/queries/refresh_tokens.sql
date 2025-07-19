-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: FindRefreshTokenByToken :one
SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens
WHERE token = $1 LIMIT 1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;