-- +goose Up
-- +goose StatementBegin
create table expenses (
    id integer primary key,

    -- time is stored as unix time
    created_at integer,
    occured_at integer,

    description text,

    -- stored as cents, not dollars
    amount integer
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table expenses;
-- +goose StatementEnd
