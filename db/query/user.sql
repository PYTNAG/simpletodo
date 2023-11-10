-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
	username, 
	hash
) VALUES (
	$1, $2
) RETURNING *;

-- name: RehashUser :one
UPDATE users
	set hash = sqlc.arg(new_hash)
WHERE id = $1
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING id, username;