package expenses

import (
	"context"
	"time"
)

type SummaryTimeRange int

const (
	AllExpenses SummaryTimeRange = iota
	ThisMonth
	CustomMonth
	ThisYear
	CustomYear
	CustomRange
)

// Service implements all of the underlying business logic.
// Things such as expenses being positive and not zero, etc.
type Service struct {
	repo Repository
}

// NewService utilizes the Repository interface defined in internal/repository.go
// This way, we never need to worry about the underlying database
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) NewExpense(ctx context.Context, occuredAt time.Time, description string, amount int64) (*Expense, error) {
	return nil, nil
}

func (s *Service) GetAllExpenses(ctx context.Context) ([]*Expense, error) {
	return nil, nil
}

func (s *Service) GetExpenseByID(ctx context.Context, id int) (*Expense, error) {
	return nil, nil
}

func (s *Service) UpdateExpense(ctx context.Context, id int, occuredAt time.Time, description string, amount int64) error {
	return nil
}

func (s *Service) DeleteExpense(ctx context.Context, id int) error {
	return nil
}

func (s *Service) SummarizeExpenses(ctx context.Context, kind SummaryTimeRange, modifier string) (string, error) {
	return "", nil
}
