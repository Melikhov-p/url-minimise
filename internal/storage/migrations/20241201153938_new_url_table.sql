-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS url(
                     short_url VARCHAR(255) UNIQUE NOT NULL,
                     original_url TEXT NOT NULL UNIQUE,
                     uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS url;
-- +goose StatementEnd
