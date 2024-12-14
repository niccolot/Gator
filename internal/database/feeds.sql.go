// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: feeds.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createFeed = `-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
	$1,
	$2,
	$3,
	$4,
    $5,
    $6
)
RETURNING id, created_at, updated_at, name, url, user_id, last_fetched_at
`

type CreateFeedParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Url       string
	UserID    uuid.UUID
}

func (q *Queries) CreateFeed(ctx context.Context, arg CreateFeedParams) (Feed, error) {
	row := q.db.QueryRowContext(ctx, createFeed,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Url,
		arg.UserID,
	)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.LastFetchedAt,
	)
	return i, err
}

const getFeedFromID = `-- name: GetFeedFromID :one
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at FROM feeds
WHERE id = $1
`

func (q *Queries) GetFeedFromID(ctx context.Context, id uuid.UUID) (Feed, error) {
	row := q.db.QueryRowContext(ctx, getFeedFromID, id)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.LastFetchedAt,
	)
	return i, err
}

const getFeedFromURL = `-- name: GetFeedFromURL :one
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at FROM feeds 
WHERE url = $1
`

func (q *Queries) GetFeedFromURL(ctx context.Context, url string) (Feed, error) {
	row := q.db.QueryRowContext(ctx, getFeedFromURL, url)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.LastFetchedAt,
	)
	return i, err
}

const getFeeds = `-- name: GetFeeds :many
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at FROM feeds
`

func (q *Queries) GetFeeds(ctx context.Context) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, getFeeds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.UserID,
			&i.LastFetchedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNextFeedToFetch = `-- name: GetNextFeedToFetch :one
SELECT feeds.id, feeds.created_at, feeds.updated_at, name, url, feeds.user_id, last_fetched_at, feed_follows.id, feed_follows.created_at, feed_follows.updated_at, feed_follows.user_id, feed_id
FROM feeds INNER JOIN feed_follows ON feeds.id = feed_follows.feed_id
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1
`

type GetNextFeedToFetchRow struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string
	Url           string
	UserID        uuid.UUID
	LastFetchedAt sql.NullTime
	ID_2          uuid.UUID
	CreatedAt_2   time.Time
	UpdatedAt_2   time.Time
	UserID_2      uuid.UUID
	FeedID        uuid.UUID
}

func (q *Queries) GetNextFeedToFetch(ctx context.Context) (GetNextFeedToFetchRow, error) {
	row := q.db.QueryRowContext(ctx, getNextFeedToFetch)
	var i GetNextFeedToFetchRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.LastFetchedAt,
		&i.ID_2,
		&i.CreatedAt_2,
		&i.UpdatedAt_2,
		&i.UserID_2,
		&i.FeedID,
	)
	return i, err
}

const markFeedFetched = `-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $2, 
	updated_at = $2
WHERE id = $1
`

type MarkFeedFetchedParams struct {
	ID            uuid.UUID
	LastFetchedAt sql.NullTime
}

func (q *Queries) MarkFeedFetched(ctx context.Context, arg MarkFeedFetchedParams) error {
	_, err := q.db.ExecContext(ctx, markFeedFetched, arg.ID, arg.LastFetchedAt)
	return err
}
