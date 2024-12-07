-- +goose Up
-- +goose StatementBegin
ALTER TABLE feeds
ADD CONSTRAINT unique_url UNIQUE (url);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE feeds
DROP CONSTRAINT unique_url;
-- +goose StatementEnd
