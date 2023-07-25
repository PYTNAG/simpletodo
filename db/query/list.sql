-- name: GetLists :many
SELECT id, header FROM lists
WHERE author = $1;

-- name: AddList :one
INSERT INTO lists (
	author, header
) VALUES (
	$1, $2
) RETURNING *;

-- name: DeleteList :exec
DELETE FROM lists
WHERE id = $1;