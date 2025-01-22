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
WITH users_posts AS (
    SELECT feed_follows.feed_id,
            feeds.name 
    FROM feed_follows 
    INNER JOIN feeds ON feed_follows.feed_id = feeds.id 
    WHERE feed_follows.user_id = $1
)
SELECT 
    posts.title,
    posts.url,
    posts.published_at,
    posts.description,
    users_posts.name AS feed_name
FROM posts
INNER JOIN users_posts ON users_posts.feed_id = posts.feed_id
ORDER BY COALESCE(posts.created_at, posts.updated_at) DESC
LIMIT $2;
