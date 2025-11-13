package handler_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/handler"
)

// mockService implementes the expenses service in order to test the handler (controller) layer.
//
// We do not need to duplicate
type mockService struct {
	lastID int
	db     map[int]*expenses.Expense

	// mutex for safety
	mux *sync.RWMutex
}

func (m *mockService) GetAllExpenses(ctx context.Context) ([]*expenses.Expense, error) {
	// return empty if
	if len(m.db) == 0 {
		return []*expenses.Expense{}, nil
	}

	// get records
	m.mux.RLock()
	defer m.mux.RUnlock()

	records := make([]*expenses.Expense, 0)
	for _, record := range m.db {
		records = append(records, record)
	}

	return records, nil
}

func (m *mockService) NewExpense(ctx context.Context, occuredAt time.Time, description string, amount int64) (*expenses.Expense, error) {
	// increment last id
	m.lastID++

	// create new mock record with last id
	id := m.lastID
	record := &expenses.Expense{
		ID:               id,
		Amount:           amount,
		ExpenseOccuredAt: occuredAt,
		RecordCreatedAt:  time.Now(),
		Description:      description,
	}

	// insert into the mock db
	m.mux.Lock()
	defer m.mux.Unlock()
	m.db[id] = record

	// return the created record
	return record, nil
}

func (m *mockService) GetExpenseByID(ctx context.Context, id int) (*expenses.Expense, error) {
	// check for id validity
	if id <= 0 {
		return nil, expenses.ErrInvalidID
	} else if id >= m.lastID {
		return nil, expenses.ErrInvalidID
	}

	// get the record
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.db[id], nil
}

func (m *mockService) UpdateExpense(ctx context.Context, id int, occuredAt time.Time, description string, amount int64) error {
	// check for id validity
	if id <= 0 {
		return expenses.ErrInvalidID
	} else if id >= m.lastID {
		return expenses.ErrInvalidID
	}

	// update record
	m.mux.Lock()
	defer m.mux.Unlock()

	// get exisiting record
	record := m.db[id]

	// update record
	record.ExpenseOccuredAt = occuredAt
	record.Description = description
	record.Amount = amount

	// insert record
	m.db[id] = record

	return nil
}

func (m *mockService) DeleteExpense(ctx context.Context, id int) error {
	// check for id validity
	if id <= 0 {
		return expenses.ErrInvalidID
	} else if id >= m.lastID {
		return expenses.ErrInvalidID
	}

	// delete record
	m.mux.Lock()
	defer m.mux.Unlock()

	delete(m.db, id)

	return nil
}

func (m *mockService) SummarizeExpenses(ctx context.Context, kind expenses.SummaryTimeRange, modifier string) (*expenses.ExpenseSummary, error) {
	// not implemented yet...
	fmt.Printf("oops not implemented...\n")

	return nil, nil
}

// setupMockService sets up the mock service for testing
func setupMockService(t *testing.T) expenses.Service {
	t.Helper()

	// create mock service
	db := make(map[int]*expenses.Expense, 0)

	// id starts at 0 because it is incremented when a record is inserted
	id := 0

	s := &mockService{
		lastID: id,
		db:     db,
		mux:    &sync.RWMutex{},
	}

	// insert 'records'

	// return setup service
	return s
}

func TestGetAllExpenses(t *testing.T) {
	testTable := []struct {
		name string
	}{
		{
			name: "valid-request",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// setup mock repo/service
			s := setupMockService(t)
			h := handler.NewExpanseHandler(s)

			// defered teardown
			// call function
			// check errors
			// check response
		})
	}
}
