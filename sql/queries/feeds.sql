-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
	$1,
	$2,
	$3,
	$4,
    $5,
    $6
)
RETURNING *;

CREATE INDEX idx_feeds_url ON feeds(url);

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedFromID :one
SELECT * FROM feeds
WHERE id = $1;

-- name: GetFeedFromURL :one
SELECT * FROM feeds 
WHERE url = $1;