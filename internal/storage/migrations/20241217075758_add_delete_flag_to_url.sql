-- +goose Up
-- +goose StatementBegin
ALTER TABLE url ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url DROP COLUMN is_deleted;
-- +goose StatementEnd
