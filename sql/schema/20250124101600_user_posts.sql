-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_posts(
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    post_id UUID NOT NULL,
    CONSTRAINT unique_user_post_pair UNIQUE (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_posts;
-- +goose StatementEnd
