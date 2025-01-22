-- +goose Up
-- +goose StatementBegin
ALTER TABLE feeds
ADD CONSTRAINT unique_name UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE feeds
DROP CONSTRAINT unique_name;
-- +goose StatementEnd