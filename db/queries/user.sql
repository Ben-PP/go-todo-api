-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;

-- name: GetAllUsers :many
SELECT id, username, is_admin, created_at
FROM users;

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash, is_admin)
VALUES ($1, $2, $3, $4)
RETURNING id, username, is_admin, created_at;

-- name: UpdateUser :one
UPDATE users
SET username = $2, is_admin = $3
WHERE id = $1
RETURNING id, username, is_admin, created_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;