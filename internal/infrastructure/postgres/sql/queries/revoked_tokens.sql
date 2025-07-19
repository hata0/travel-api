-- name: CreateRevokedToken :exec
INSERT INTO revoked_tokens (id, token_jti, expires_at, revoked_at)
VALUES ($1, $2, $3, $4);

-- name: GetRevokedTokenByJTI :one
SELECT id, token_jti, expires_at, revoked_at FROM revoked_tokens
WHERE token_jti = $1 LIMIT 1;