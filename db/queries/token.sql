-- name: CreateJwtToken :exec
INSERT INTO jwt_tokens (jti, user_id, family, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetJwtTokenByJti :one
SELECT *
FROM jwt_tokens
WHERE jti = $1;


-- name: UseJwtToken :exec
UPDATE jwt_tokens
SET is_used = TRUE
WHERE jti = $1;

-- name: DeleteJwtTokenByFamily :execrows
DELETE FROM jwt_tokens
WHERE family = $1;

-- name: DeleteJwtTokenByUserIdExcludeFamily :exec
DELETE FROM jwt_tokens
WHERE user_id = $1 AND family != $2;

-- name: DeleteJwtTokensByUserId :exec
DELETE FROM jwt_tokens
WHERE user_id = $1;