-- +goose Up
-- +goose StatementBegin
CREATE TABLE task (
    id uuid PRIMARY KEY,
    subject varchar(255),
    description text,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_date date
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP table task;
-- +goose StatementEnd
