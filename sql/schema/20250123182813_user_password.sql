-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN hashed_password TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN hashed_password;
-- +goose StatementEnd