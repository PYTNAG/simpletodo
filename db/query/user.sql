-- name: GetUser :one
SELECT id, username FROM users
WHERE username = $1 AND hash = $2 LIMIT 1;

-- name: AddUser :one
INSERT INTO users (
	username, hash
) VALUES (
	$1, $2
) RETURNING *;

-- name: RehashUser :one
UPDATE users
	set hash = sqlc.arg(new_hash)
WHERE id = $1 and hash = sqlc.arg(old_hash)
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1 AND hash = $2
RETURNING id, username;