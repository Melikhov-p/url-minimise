-- +goose Up
-- +goose StatementBegin
ALTER TABLE url
    ADD COLUMN user_id INTEGER,
    ADD CONSTRAINT fk_user
        FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url
    DROP CONSTRAINT fk_user;
ALTER TABLE url
    DROP COLUMN user_id;
-- +goose StatementEnd
