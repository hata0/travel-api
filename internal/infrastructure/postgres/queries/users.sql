-- name: CreateUser :exec
INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: FindUserByEmail :one
SELECT id, username, email, password_hash, created_at, updated_at FROM users
WHERE email = $1;

-- name: FindUserByUsername :one
SELECT id, username, email, password_hash, created_at, updated_at FROM users
WHERE username = $1;

-- name: FindUser :one
SELECT id, username, email, password_hash, created_at, updated_at FROM users
WHERE id = $1;
