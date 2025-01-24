-- name: BookmarkPost :one
INSERT INTO user_posts (id, created_at, user_id, post_id)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetBookmarkedPostsForUser :many
SELECT * FROM user_posts
WHERE user_id = $1;