-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
	$1,
	$2,
	$3,
	$4
)
RETURNING *;

CREATE INDEX idx_users_name ON users(name);

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users 
WHERE name = $1;

-- name: GetuserFromID :one
SELECT * FROM users
WHERE id = $1;

-- name: Reset :exec
DELETE FROM users;