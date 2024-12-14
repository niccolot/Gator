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

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $2, 
	updated_at = $2
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds INNER JOIN feed_follows ON feeds.id = feed_follows.feed_id
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;