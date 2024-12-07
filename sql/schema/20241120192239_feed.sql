-- +goose Up
-- +goose StatementBegin
CREATE TABLE feeds(
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    user_id UUID NOT NULL
);

ALTER TABLE feeds
ADD CONSTRAINT fk_user
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE feeds
DROP CONSTRAINT fk_user;

DROP TABLE feeds;
-- +goose StatementEnd