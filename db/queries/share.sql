-- name: CreateShare :one
INSERT INTO list_shares (list_id, user_id)
VALUES ($1, $2)
RETURNING list_id, user_id;

-- name: DeleteShare :exec
DELETE FROM list_shares
WHERE list_id = $1 AND user_id = $2;