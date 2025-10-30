-- name: CreateShare :one
INSERT INTO list_shares (list_id, user_id)
VALUES ($1, $2)
RETURNING list_id, user_id;