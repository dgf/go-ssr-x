-- +goose Up
-- +goose StatementBegin
CREATE TABLE task (
    id uuid NOT NULL,
    subject varchar(255),
    description text,
    created_at timestamp NOT NULL DEFAULT NOW(),
    due_date date,
    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE task;
-- +goose StatementEnd
