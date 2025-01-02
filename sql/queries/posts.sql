-- name: CreatePost :one
INSERT INTO posts (
    id, 
    created_at,
    updated_at,
    title,
    url,
    description,
    published_at,
    feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
) RETURNING *;

CREATE INDEX idx_posts_feed_id ON posts(feed_id);

CREATE INDEX idx_posts_updated_at ON posts(updated_at);

CREATE INDEX idx_posts_title ON posts(title);

-- name: GetPostFromTitle :one
SELECT * FROM posts
WHERE title = $1;

-- name: UpdatePost :exec
UPDATE posts
SET updated_at = $2
WHERE id = $1;

-- name: GetPostsForUser :many
SELECT * FROM posts
ORDER BY COALESCE(posts.published_at, posts.updated_at) DESC
LIMIT $1;
