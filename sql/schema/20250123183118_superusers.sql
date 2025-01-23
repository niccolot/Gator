-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN is_superuser BOOL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN is_superuser;
-- +goose StatementEnd