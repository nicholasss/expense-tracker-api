package handler_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
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
		RecordCreatedAt:  time.Unix(0, 0), // recorded time doesnt matter :) tested elsewhere
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
	records := []expenses.Expense{
		{
			Amount:           1999,
			ExpenseOccuredAt: time.Unix(1763398641, 0),
			Description:      "movie tickets",
		},
		{
			Amount:           28089,
			ExpenseOccuredAt: time.Unix(1763402231, 0),
			Description:      "big fancy dinner",
		},
		{
			Amount:           940,
			ExpenseOccuredAt: time.Unix(1763405881, 0),
			Description:      "parking payment",
		},
	}

	for _, record := range records {
		_, err := s.NewExpense(t.Context(), record.ExpenseOccuredAt, record.Description, record.Amount)
		if err != nil {
			t.Fatalf("unable to insert records into mock database due to: %s", err)
		}
	}

	// return setup service
	return s
}

func TestGetAllExpenses(t *testing.T) {
	testTable := []struct {
		name        string
		wantCode    int
		wantRecords []*expenses.Expense
		wantBody    string
		wantHeaders map[string]string
	}{
		{
			name:     "valid-request",
			wantCode: 200,
			wantRecords: []*expenses.Expense{
				{
					ID:               1,
					Amount:           1999,
					ExpenseOccuredAt: time.Unix(1763398641, 0),
					Description:      "movie tickets",
				},
				{
					ID:               2,
					Amount:           28089,
					ExpenseOccuredAt: time.Unix(1763402231, 0),
					Description:      "big fancy dinner",
				},
				{
					ID:               3,
					Amount:           940,
					ExpenseOccuredAt: time.Unix(1763405881, 0),
					Description:      "parking payment",
				},
			},
			wantBody:    `[{"id":1,"created_at":"1970-01-01T00:00:00Z","occured_at":"2025-11-17T16:57:21Z","description":"movie tickets","amount":1999},{"id":2,"created_at":"1970-01-01T00:00:00Z","occured_at":"2025-11-17T17:57:11Z","description":"big fancy dinner","amount":28089},{"id":3,"created_at":"1970-01-01T00:00:00Z","occured_at":"2025-11-17T18:58:01Z","description":"parking payment","amount":940}]`,
			wantHeaders: map[string]string{"Content-Type": "application/json"},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// setup mock repo/service
			service := setupMockService(t)
			handler := handler.NewExpanseHandler(service)

			// test request
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/expenses", http.NoBody)

			// response recorder
			recorder := httptest.NewRecorder()

			// call handler
			handler.GetAllExpenses(recorder, request)
			gotResp := recorder.Result()

			gotBody, err := io.ReadAll(gotResp.Body)
			if err != nil {
				t.Fatalf("cannot read response body due to: %s", err)
			}

			// defering body closure
			defer func() {
				err := gotResp.Body.Close()
				if err != nil {
					t.Fatalf("unable to close test response due to: %s", err)
				}
			}()

			// getting headers
			gotHeaders := gotResp.Header.Clone()

			// check response code
			if gotResp.StatusCode != testCase.wantCode {
				t.Errorf("got status HTTP %d, wanted status HTTP %d", gotResp.StatusCode, testCase.wantCode)
			}

			// check headers
			if gotHeaders == nil {
				t.Errorf("response did not have headers")
			} else {
				for wantHeaderKey, wantHeaderVal := range testCase.wantHeaders {

					// checking each header for existence and its values
					gotHeaderValsForKey, exists := gotHeaders[wantHeaderKey]
					if !exists {
						t.Errorf("got nothing for header %q want value %q", wantHeaderKey, wantHeaderVal)
						continue
					}

					// checking for header values of the key provided
					if !slices.Contains(gotHeaderValsForKey, wantHeaderVal) {
						t.Errorf("under header %q, got values of %+q, wanted value %q", wantHeaderKey, gotHeaderValsForKey, wantHeaderVal)
					}

				}
			}

			// check response body
			if string(gotBody) != testCase.wantBody {
				t.Errorf("got body: \n'%s'\nwant body: \n'%s'\n", gotBody, testCase.wantBody)
			}
		})
	}
}
