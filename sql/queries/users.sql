-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, hashed_password, is_superuser)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6
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

-- name: ResetUsers :exec
DELETE FROM users;

-- name: UpdateToSuper :exec
UPDATE users
SET is_superuser = TRUE
WHERE id = $1;

-- name: ChangePassword :exec
UPDATE users
SET hashed_password = $2
WHERE id = $1;