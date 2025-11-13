package expenses

import (
	"context"
	"time"
)

// Service defines an interface for the business layer of the API.
//
// This is primarily implemented for easier mocking for testing.
type Service interface {
	NewExpense(ctx context.Context, occuredAt time.Time, description string, amount int64) (*Expense, error)

	GetAllExpenses(ctx context.Context) ([]*Expense, error)

	GetExpenseByID(ctx context.Context, id int) (*Expense, error)

	UpdateExpense(ctx context.Context, id int, occuredAt time.Time, description string, amount int64) error

	DeleteExpense(ctx context.Context, id int) error

	SummarizeExpenses(ctx context.Context, kind SummaryTimeRange, modifier string) (*ExpenseSummary, error)
}
