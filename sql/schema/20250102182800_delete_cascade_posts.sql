-- +goose Up
-- +goose StatementBegin
ALTER TABLE posts
ADD CONSTRAINT fk_feed
FOREIGN KEY (feed_id)
REFERENCES feeds(id)
ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE posts
DROP CONSTRAINT fk_feed;
-- +goose StatementEnd
