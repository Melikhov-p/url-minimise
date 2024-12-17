-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "user"(
    id SERIAL PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "user";
-- +goose StatementEnd
