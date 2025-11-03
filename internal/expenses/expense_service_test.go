package expenses_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/sqlite"
)

// mockRepository implements the Respository interface to test the service layer
// we are not testing the repository layer, so we just need to ACT like we are performing the action
// and make sure that the actions within the service layer are performing as expected
type mockRepository struct {
	lastID int
	db     map[int]*expenses.Expense

	// mutex for safety
	mux *sync.RWMutex
}

// get one expense record by ID
func (r *mockRepository) GetByID(ctx context.Context, id int) (*expenses.Expense, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	// get from map of records
	record, ok := r.db[id]
	if !ok {
		return nil, &sqlite.QueryError{Query: "mocked for test", Err: sql.ErrNoRows}
	}

	return record, nil
}

// get all expenses
func (r *mockRepository) GetAll(ctx context.Context) ([]*expenses.Expense, error) {
	// check if no records in repository
	if r.lastID == 0 {
		return nil, &sqlite.QueryError{Query: "mocked for test", Err: sql.ErrNoRows}
	}

	// lock mux
	r.mux.RLock()
	defer r.mux.RUnlock()

	// read all records into slice
	records := make([]*expenses.Expense, 0)
	for i := range r.lastID {
		record, ok := r.db[i]

		// only append if not deleted
		if ok {
			records = append(records, record)
		}
	}

	// return slice
	return records, nil
}

// create a new expense
func (r *mockRepository) Create(ctx context.Context, exp *expenses.Expense) (*expenses.Expense, error) {
	// check for nil exp pointer
	if exp == nil {
		return nil, sqlite.ErrNilPointer
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	// get new id
	r.lastID += 1
	newID := r.lastID

	// prepare record
	exp.ID = newID
	exp.RecordCreatedAt = time.Now()

	// insert record into map
	r.db[newID] = exp

	return exp, nil
}

// update an existing expense
func (r *mockRepository) Update(ctx context.Context, exp *expenses.Expense) error {
	// check for nil exp pointer
	if exp == nil {
		return sqlite.ErrNilPointer
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	// make sure id exists
	_, exists := r.db[exp.ID]
	if !exists {
		return sqlite.ErrNoRowsUpdated
	}

	// perform update
	r.db[exp.ID] = exp

	return nil
}

// delete an exisiting expense
func (r *mockRepository) Delete(ctx context.Context, id int) error {
	// lock mux
	r.mux.Lock()
	defer r.mux.Unlock()

	// test if empty, and return ErrNoRowsDeleted
	_, exists := r.db[id]
	if r.lastID == 0 || !exists {
		return sqlite.ErrNoRowsDeleted
	}

	// remove record because it exists
	delete(r.db, id)
	return nil
}

// setupTestRepo sets up a mock repository layer in order to test the service layer
func setupTestRepo(t *testing.T) expenses.Repository {
	t.Helper()

	// setup repo
	repo := &mockRepository{
		lastID: 0,
		db:     make(map[int]*expenses.Expense, 10),
		mux:    &sync.RWMutex{},
	}

	// list out records to load
	recordsToLoad := []*expenses.Expense{
		{
			Amount:           8929,
			ExpenseOccuredAt: time.Unix(1760574600, 0),
			Description:      "dinner out with friends",
		},
		{
			Amount:           7800,
			ExpenseOccuredAt: time.Unix(1760792400, 0),
			Description:      "coffee for office breakfast",
		},
		{
			Amount:           4810,
			ExpenseOccuredAt: time.Unix(1760877900, 0),
			Description:      "bagels for office breakfast",
		},
		{
			Amount:           31800,
			ExpenseOccuredAt: time.Unix(1761160500, 0),
			Description:      "new CAT5 cabling for office",
		},
		{
			Amount:           74100,
			ExpenseOccuredAt: time.Unix(1761404400, 0),
			Description:      "replacing breaks on company prius",
		},
		{
			Amount:           289,
			ExpenseOccuredAt: time.Unix(1761670800, 0),
			Description:      "soda from maplefields",
		},
	}

	// load in records
	for _, record := range recordsToLoad {
		_, err := repo.Create(t.Context(), record)
		if err != nil {
			t.Fatalf("Unable to setup test repo due to: %v", err)
		}
	}

	// return :)
	return repo
}

func checkExpenseEquality(t *testing.T, got, want *expenses.Expense) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("expenses.ID does not match. got: %v, want: %v", got.ID, want.ID)
	}
	if got.ExpenseOccuredAt != want.ExpenseOccuredAt {
		t.Errorf("expenses.ExpenseOccuredAt does not match. got: %v, want: %v", got.ExpenseOccuredAt, want.ExpenseOccuredAt)
	}
	if got.Description != want.Description {
		t.Errorf("expenses.Description does not match. got: %v, want: %v", got.Description, want.Description)
	}
	if got.Amount != want.Amount {
		t.Errorf("expenses.Amount does not match. got: %v, want: %v", got.Amount, want.Amount)
	}

	// not checking created at for now...
}

// TestNewExpense tests the service layer for expenses
func TestNewExpense(t *testing.T) {
	testTable := []struct {
		name             string
		inputOccuredAt   time.Time
		inputDescription string
		inputAmount      int64
		wantRecord       *expenses.Expense
		expectError      bool
		wantError        error
	}{
		{
			name:             "valid-expense-creation-b",
			inputOccuredAt:   time.Unix(1761677891, 0),
			inputDescription: "new coffee beans for espresso",
			inputAmount:      2149,
			wantRecord: &expenses.Expense{
				ID:               7,
				Amount:           2149,
				ExpenseOccuredAt: time.Unix(1761677891, 0),
				Description:      "new coffee beans for espresso",
			},
			expectError: false,
			wantError:   nil,
		},
		{
			name:             "valid-expense-creation-a",
			inputOccuredAt:   time.Unix(1761721091, 0),
			inputDescription: "new cat food for bodega cat",
			inputAmount:      3499,
			wantRecord: &expenses.Expense{
				ID:               7,
				Amount:           3499,
				ExpenseOccuredAt: time.Unix(1761721091, 0),
				Description:      "new cat food for bodega cat",
			},
			expectError: false,
			wantError:   nil,
		},
		{
			name:             "invalid-occured-at-time",
			inputOccuredAt:   time.Unix(0, 0),
			inputDescription: "new cat food for bodega cat",
			inputAmount:      3499,
			wantRecord:       nil,
			expectError:      true,
			wantError:        expenses.ErrInvalidOccuredAtTime,
		},
		{
			name:             "invalid-amount-zero",
			inputOccuredAt:   time.Unix(1761721091, 0),
			inputDescription: "new cat food for bodega cat",
			inputAmount:      0,
			wantRecord:       nil,
			expectError:      true,
			wantError:        expenses.ErrInvalidAmount,
		},
		{
			name:             "invalid-amount-negative",
			inputOccuredAt:   time.Unix(1761721091, 0),
			inputDescription: "new cat food for bodega cat",
			inputAmount:      -2,
			wantRecord:       nil,
			expectError:      true,
			wantError:        expenses.ErrInvalidAmount,
		},
		{
			name:             "invalid-empty-description",
			inputOccuredAt:   time.Unix(1761721091, 0),
			inputDescription: "",
			inputAmount:      3499,
			wantRecord:       nil,
			expectError:      true,
			wantError:        expenses.ErrInvalidDescription,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			repo := setupTestRepo(t)
			serv := expenses.NewService(repo)

			// call function
			gotRecord, gotErr := serv.NewExpense(t.Context(),
				testCase.inputOccuredAt, testCase.inputDescription, testCase.inputAmount,
			)

			// test for expecting error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("NewExpense() got error: '%v', expected error: '%v'", gotErr, testCase.wantError)
			}

			// test on error if expected
			if gotErr != nil {
				if !errors.Is(gotErr, testCase.wantError) {
					t.Errorf("got error: %v, want error: %v", gotErr, testCase.wantError)
				}
			}

			// test on returned expense
			if !testCase.expectError && gotRecord != nil {
				checkExpenseEquality(t, gotRecord, testCase.wantRecord)
			}
		})
	}
}

func TestGetAllExpenses(t *testing.T) {
	testTable := []struct {
		name        string
		wantRecords []*expenses.Expense
		expectError bool
		wantError   error
	}{
		{
			name:        "valid-all-records",
			expectError: false,
			wantError:   nil,
			wantRecords: []*expenses.Expense{
				{
					ID:               1,
					Amount:           8929,
					ExpenseOccuredAt: time.Unix(1760574600, 0),
					Description:      "dinner out with friends",
				},
				{
					ID:               2,
					Amount:           7800,
					ExpenseOccuredAt: time.Unix(1760792400, 0),
					Description:      "coffee for office breakfast",
				},
				{
					ID:               3,
					Amount:           4810,
					ExpenseOccuredAt: time.Unix(1760877900, 0),
					Description:      "bagels for office breakfast",
				},
				{
					ID:               4,
					Amount:           31800,
					ExpenseOccuredAt: time.Unix(1761160500, 0),
					Description:      "new CAT5 cabling for office",
				},
				{
					ID:               5,
					Amount:           74100,
					ExpenseOccuredAt: time.Unix(1761404400, 0),
					Description:      "replacing breaks on company prius",
				},
				{
					ID:               6,
					Amount:           289,
					ExpenseOccuredAt: time.Unix(1761670800, 0),
					Description:      "soda from maplefields",
				},
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			repo := setupTestRepo(t)
			serv := expenses.NewService(repo)

			// call function
			gotRecords, gotErr := serv.GetAllExpenses(t.Context())

			// checking if we expect an error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("GetAll() got error: '%v', expected error: '%v'", gotErr, testCase.wantError)
			}

			// checking error type if its not nil
			if gotErr != nil {
				if !errors.Is(gotErr, testCase.wantError) {
					t.Errorf("got error: %v, want error: %v", gotErr, testCase.wantError)
				}
			}

			// checking result
			if !testCase.expectError && gotRecords != nil {
				for i, gotRecord := range gotRecords {

					t.Logf("Record %d mismatch", i+1)
					checkExpenseEquality(t, gotRecord, testCase.wantRecords[i])
				}
			}
		})
	}
}

func TestGetExpenseByID(t *testing.T) {
	testTable := []struct {
		name        string
		inputID     int
		expectError bool
		wantError   error
		wantRecord  *expenses.Expense
	}{
		{
			name:        "valid-first-expense",
			inputID:     1,
			expectError: false,
			wantError:   nil,
			wantRecord: &expenses.Expense{
				ID:               1,
				Amount:           8929,
				ExpenseOccuredAt: time.Unix(1760574600, 0),
				Description:      "dinner out with friends",
			},
		},
		{
			name:        "valid-third-expense",
			inputID:     3,
			expectError: false,
			wantError:   nil,
			wantRecord: &expenses.Expense{
				ID:               3,
				Amount:           4810,
				ExpenseOccuredAt: time.Unix(1760877900, 0),
				Description:      "bagels for office breakfast",
			},
		},
		{
			name:        "invalid-zero-id",
			inputID:     0,
			expectError: true,
			wantError:   expenses.ErrInvalidID,
			wantRecord:  nil,
		},
		{
			name:        "invalid-negative-id",
			inputID:     -2,
			expectError: true,
			wantError:   expenses.ErrInvalidID,
			wantRecord:  nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			repo := setupTestRepo(t)
			serv := expenses.NewService(repo)

			// call function
			gotRecord, gotErr := serv.GetExpenseByID(t.Context(), testCase.inputID)

			// checking if we expect an error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("GetByID(%q) got error: '%v', expected error: '%v'", testCase.inputID, gotErr, testCase.wantError)
			}

			// checking error type if its not nil
			if gotErr != nil {
				if !errors.Is(gotErr, testCase.wantError) {
					t.Errorf("got error: %v, want error: %v", gotErr, testCase.wantError)
				}
			}

			// checking result
			if !testCase.expectError && gotRecord != nil {
				checkExpenseEquality(t, gotRecord, testCase.wantRecord)
			}
		})
	}
}

func TestUpdateExpense(t *testing.T) {
	testTable := []struct {
		name             string
		inputID          int
		inputOccuredAt   time.Time
		inputDescription string
		inputAmount      int64
		expectError      bool
		wantError        error
	}{
		{
			name:             "valid-update-first-expense's-description",
			inputID:          1,
			inputOccuredAt:   time.Unix(1760574600, 0),
			inputDescription: "dinner out with friends, not family",
			inputAmount:      8929,
			expectError:      false,
			wantError:        nil,
		},
		{
			name:             "valid-update-last-expense's-amount",
			inputID:          6,
			inputOccuredAt:   time.Unix(1761670800, 0),
			inputDescription: "soda from maplefields",
			inputAmount:      189,
			expectError:      false,
			wantError:        nil,
		},
		{
			name:             "valid-update-third-expense's-occured-at-time",
			inputID:          3,
			inputOccuredAt:   time.Unix(1760877820, 0),
			inputDescription: "bagels for office breakfast",
			inputAmount:      4810,
			expectError:      false,
			wantError:        nil,
		},
		{
			name:             "invalid-zero-id",
			inputID:          0,
			inputOccuredAt:   time.Unix(1760574600, 0),
			inputDescription: "dinner out with friends, not family",
			inputAmount:      8929,
			expectError:      true,
			wantError:        expenses.ErrInvalidID,
		},
		{
			name:             "invalid-negative-id",
			inputID:          -2,
			inputOccuredAt:   time.Unix(1760574600, 0),
			inputDescription: "dinner out with friends, not family",
			inputAmount:      8929,
			expectError:      true,
			wantError:        expenses.ErrInvalidID,
		},
		{
			name:             "invalid-occured-at-time",
			inputID:          3,
			inputOccuredAt:   time.Unix(0, 0),
			inputDescription: "bagels for office breakfast",
			inputAmount:      4810,
			expectError:      true,
			wantError:        expenses.ErrInvalidOccuredAtTime,
		},
		{
			name:             "invalid-empty-description",
			inputID:          3,
			inputOccuredAt:   time.Unix(1760877900, 0),
			inputDescription: "",
			inputAmount:      4810,
			expectError:      true,
			wantError:        expenses.ErrInvalidDescription,
		},
		{
			name:             "invalid-amount-zero",
			inputID:          3,
			inputOccuredAt:   time.Unix(1760877900, 0),
			inputDescription: "bagels for office breakfast",
			inputAmount:      0,
			expectError:      true,
			wantError:        expenses.ErrInvalidAmount,
		},
		{
			name:             "invalid-amount-negative",
			inputID:          3,
			inputOccuredAt:   time.Unix(1760877900, 0),
			inputDescription: "bagels for office breakfast",
			inputAmount:      -2,
			expectError:      true,
			wantError:        expenses.ErrInvalidAmount,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			repo := setupTestRepo(t)
			serv := expenses.NewService(repo)

			// call function
			gotErr := serv.UpdateExpense(t.Context(),
				testCase.inputID, testCase.inputOccuredAt, testCase.inputDescription, testCase.inputAmount)

			// checking if we expect an error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("Update() got error: '%v', expected error: '%v'", gotErr, testCase.wantError)
			}

			// checking error type if its not nil
			if gotErr != nil {
				if !errors.Is(gotErr, testCase.wantError) {
					t.Errorf("got error: %v, want error: %v", gotErr, testCase.wantError)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testTable := []struct {
		name        string
		inputID     int
		expectError bool
		wantError   error
	}{
		{
			name:        "valid-delete-first",
			inputID:     1,
			expectError: false,
			wantError:   nil,
		},
		{
			name:        "valid-delete-last",
			inputID:     6,
			expectError: false,
			wantError:   nil,
		},
		{
			name:        "invalid-zero-id",
			inputID:     0,
			expectError: true,
			wantError:   expenses.ErrInvalidID,
		},
		{
			name:        "invalid-negative-id",
			inputID:     -2,
			expectError: true,
			wantError:   expenses.ErrInvalidID,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			repo := setupTestRepo(t)
			serv := expenses.NewService(repo)

			// call function
			gotErr := serv.DeleteExpense(t.Context(), testCase.inputID)

			// checking if we expect an error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("Delete() got error: '%v', expected error: '%v'", gotErr, testCase.wantError)
			}

			// checking error type if its not nil
			if gotErr != nil {
				if !errors.Is(gotErr, testCase.wantError) {
					t.Errorf("got error: %v, want error: %v", gotErr, testCase.wantError)
				}
			}
		})
	}
}

func TestSummarizeExpenses(t *testing.T) {
	testTable := []struct {
		name          string
		inputKind     expenses.SummaryTimeRange
		inputModifier string
		wantSummary   *expenses.ExpenseSummary
		expectError   bool
		wantError     error
	}{
		//
		// this month
		{
			name:          "valid-this-month-summary",
			inputKind:     expenses.ThisMonth,
			inputModifier: "",
			wantSummary: &expenses.ExpenseSummary{
				SummaryTimeRange: "This Month",
				Total:            127439,
			},
			expectError: false,
			wantError:   nil,
		},
		//
		// custom months
		{
			name:          "valid-october-month-summary",
			inputKind:     expenses.CustomMonth,
			inputModifier: "2025-10",
			wantSummary: &expenses.ExpenseSummary{
				SummaryTimeRange: "Custom Month: October of 2025",
				Total:            127439,
			},
			expectError: false,
			wantError:   nil,
		},
		{
			name:          "valid-september-month-summary",
			inputKind:     expenses.CustomMonth,
			inputModifier: "2025-09",
			wantSummary: &expenses.ExpenseSummary{
				SummaryTimeRange: "Custom Month: September of 2025",
				Total:            0,
			},
			expectError: false,
			wantError:   nil,
		},
		{
			name:          "invalid-empty-month-modifier-summary",
			inputKind:     expenses.CustomMonth,
			inputModifier: "",
			wantSummary:   nil,
			expectError:   true,
			wantError: &expenses.ErrInvalidTime{
				ProvidedTime: "",
			},
		},
		{
			name:          "invalid-nonexistent-month-modifier-summary",
			inputKind:     expenses.CustomMonth,
			inputModifier: "2025-13",
			wantSummary:   nil,
			expectError:   true,
			wantError: &expenses.ErrInvalidTime{
				ProvidedTime: "2025-13",
			},
		},
		{
			name:          "invalid-pre-unix-epoch-month-modifier-summary",
			inputKind:     expenses.CustomMonth,
			inputModifier: "1969-09",
			wantSummary:   nil,
			expectError:   true,
			wantError: &expenses.ErrInvalidTime{
				ProvidedTime: "1969-09",
			},
		},
		//
		// this year
		{
			name:          "valid-this-year-summary",
			inputKind:     expenses.ThisYear,
			inputModifier: "",
			wantSummary: &expenses.ExpenseSummary{
				SummaryTimeRange: "This Year",
				Total:            127439,
			},
			expectError: false,
			wantError:   nil,
		},
		//
		// custom year
		{
			name:          "valid-two-years-ago-summary",
			inputKind:     expenses.CustomYear,
			inputModifier: "2023",
			wantSummary: &expenses.ExpenseSummary{
				SummaryTimeRange: "Custom Year: 2023",
				Total:            0,
			},
			expectError: false,
			wantError:   nil,
		},
		{
			name:          "invalid-empty-modifier-custom-year-summary",
			inputKind:     expenses.CustomYear,
			inputModifier: "",
			wantSummary:   nil,
			expectError:   true,
			wantError: &expenses.ErrInvalidTime{
				ProvidedTime: "",
			},
		},
		{
			name:          "invalid-pre-unix-epoch-modifier-custom-year-summary",
			inputKind:     expenses.CustomYear,
			inputModifier: "1969",
			wantSummary:   nil,
			expectError:   true,
			wantError: &expenses.ErrInvalidTime{
				ProvidedTime: "1969",
			},
		},
		// TODO:
		// custom month-year range
		// ... not implemented yet
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			repo := setupTestRepo(t)
			serv := expenses.NewService(repo)

			// call function to test
			gotSummary, gotErr := serv.SummarizeExpenses(t.Context(), testCase.inputKind, testCase.inputModifier)

			// checking if we got an error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("SummarizeExpenses() got error: '%v', expected error: '%v'", gotErr, testCase.wantError)

				// checking recieved error type
			} else if gotErr != nil {

				// checking for expenses.ErrInvalidTime error
				var wantErr *expenses.ErrInvalidTime
				if errors.As(gotErr, &wantErr) {
					// optionally can also have `wantErr.ProvidedTime != testCase.wantError.(*expenses.ErrInvalidTime).ProvidedTime`
					if wantErr.ProvidedTime != testCase.inputModifier {
						t.Errorf("got error: %v, want error: %v", gotErr, testCase.wantError)
					}

					// we got other error
				} else {
					t.Errorf("got error: %v, want error type: %T", gotErr, testCase.wantError)
				}
			}

			// checking the summary
			if gotSummary != nil {
				if gotSummary.Total != testCase.wantSummary.Total && gotSummary.SummaryTimeRange != testCase.wantSummary.SummaryTimeRange {
					t.Errorf("Expense summary does not match. got: %+v, want: %+v", gotSummary, testCase.wantSummary)
				}
			}
		})
	}
}
