-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS delete_task(
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(255),
    user_id INTEGER,
    status VARCHAR(50)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS delete_task;
-- +goose StatementEnd
