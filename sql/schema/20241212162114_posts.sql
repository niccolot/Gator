-- +goose Up
-- +goose StatementBegin
CREATE TABLE posts (
    id UUID PRIMARY KEY UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT,
    url TEXT NOT NULL,
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE posts;
-- +goose StatementEnd
