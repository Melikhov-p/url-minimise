-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS secret_key(
  id SERIAL PRIMARY KEY,
  key VARCHAR(255)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS secret_key;
-- +goose StatementEnd
