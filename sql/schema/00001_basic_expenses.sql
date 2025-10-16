-- +goose Up
-- +goose StatementBegin
create table expenses (
    expenses_id integer primary key,

    -- time is stored as unix time
    expense_created_at integer,
    expense_occured_at integer,

    description text,

    -- stored as cents, not dollars
    amount integer
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table expenses;
-- +goose StatementEnd
