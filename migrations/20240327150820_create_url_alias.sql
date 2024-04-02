-- +goose Up
-- +goose StatementBegin
CREATE TABLE url_alias(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    alias TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL);
CREATE INDEX alias_idx ON url_alias(alias);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS url_alias;
-- +goose StatementEnd
