package sqlite_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

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

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// create the in memory
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to setup in-memory sqlite database: %v", err)
	}

	// create the table
	createQuery := `
  CREATE TABLE
    expenses (
      id INTEGER PRIMARY KEY,
      created_at INTEGER,
      occured_at INTEGER,
      description TEXT,
      amount INTEGER
    );`
	_, err = db.Exec(createQuery)
	if err != nil {
		t.Fatalf("unable to create table: %v", err)
	}

	// insert data for testing
	insertQuery := `
  INSERT INTO
    expenses
      (
        created_at,
        occured_at,
        description,
        amount
      )
  VALUES
    (
      unixepoch(),
      1761231600,
      "new hairdryer",
      11999
    ),
    (
      unixepoch(),
      1761148800,
      "oat breakfast",
      1399
    ),
    (
      unixepoch(),
      1761073200,
      "cab to train station",
      2700
    ),
    (
      unixepoch(),
      1761001200,
      "late dinner with client",
      6289
    ),
    (
      unixepoch(),
      1760882400,
      "cab to lunch",
      2560
    ),
    (
      unixepoch(),
      1760810400,
      "new coffee machine for headquarters",
      18988
    );`

	_, err = db.Exec(insertQuery)
	if err != nil {
		t.Fatalf("unable to insert test data: %v", err)
	}

	return db
}

func TestGetByID(t *testing.T) {
	testTable := []struct {
		name        string
		inputID     int
		expectError bool
		wantError   error
		wantRecord  *expenses.Expense
	}{
		{
			name:        "valid-first-record-by-id",
			inputID:     1,
			expectError: false,
			wantError:   nil,
			wantRecord: &expenses.Expense{
				ID:               1,
				Amount:           11999,
				ExpenseOccuredAt: time.Unix(1761231600, 0),
				Description:      "new hairdryer",
			},
		},
		{
			name:        "valid-second-record-by-id",
			inputID:     2,
			expectError: false,
			wantError:   nil,
			wantRecord: &expenses.Expense{
				ID:               2,
				Amount:           1399,
				ExpenseOccuredAt: time.Unix(1761148800, 0),
				Description:      "oat breakfast",
			},
		},
		{
			name:        "invalid-bad-id",
			inputID:     0,
			expectError: true,
			wantError:   sql.ErrNoRows,
			wantRecord:  nil,
		},
		{
			name:        "invalid-id-does-not-exist",
			inputID:     18,
			expectError: true,
			wantError:   sql.ErrNoRows,
			wantRecord:  nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := sqlite.NewSqliteRepository(db)

			// defer teardown
			defer func() {
				err := db.Close()
				if err != nil {
					t.Errorf("unable to close connection to in-memory sqlite database: %v", err)
				}
			}()

			// calling the function
			gotRecord, gotErr := repo.GetByID(context.Background(), testCase.inputID)

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

// TestGetAll does not test whether the database is empty
func TestGetAll(t *testing.T) {
	testTable := []struct {
		name        string
		expectError bool
		wantError   error
		wantRecords []*expenses.Expense
	}{
		{
			name:        "valid-all-records",
			expectError: false,
			wantError:   nil,
			wantRecords: []*expenses.Expense{
				{
					ID:               1,
					Amount:           11999,
					ExpenseOccuredAt: time.Unix(1761231600, 0),
					Description:      "new hairdryer",
				},
				{
					ID:               2,
					Amount:           1399,
					ExpenseOccuredAt: time.Unix(1761148800, 0),
					Description:      "oat breakfast",
				},
				{
					ID:               3,
					Amount:           2700,
					ExpenseOccuredAt: time.Unix(1761073200, 0),
					Description:      "cab to train station",
				},
				{
					ID:               4,
					Amount:           6289,
					ExpenseOccuredAt: time.Unix(1761001200, 0),
					Description:      "late dinner with client",
				},
				{
					ID:               5,
					Amount:           2560,
					ExpenseOccuredAt: time.Unix(1760882400, 0),
					Description:      "cab to lunch",
				},
				{
					ID:               6,
					Amount:           18988,
					ExpenseOccuredAt: time.Unix(1760810400, 0),
					Description:      "new coffee machine for headquarters",
				},
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := sqlite.NewSqliteRepository(db)

			// defer teardown
			defer func() {
				err := db.Close()
				if err != nil {
					t.Errorf("unable to close connection to in-memory sqlite database: %v", err)
				}
			}()

			// calling the function
			gotRecords, gotErr := repo.GetAll(context.Background())

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

func TestCreate(t *testing.T) {
	// the in memory database is setup for each individual test case,
	// so the newly created record will always be ID = 7

	testTable := []struct {
		name        string
		inputRecord *expenses.Expense
		expectError bool
		wantError   error
		wantRecord  *expenses.Expense
	}{
		{
			name: "valid-first-full-record",
			inputRecord: &expenses.Expense{
				Amount:           229,
				ExpenseOccuredAt: time.Unix(1761249149, 0),
				Description:      "new altoids",
			},
			expectError: false,
			wantError:   nil,
			wantRecord: &expenses.Expense{
				ID:               7,
				Amount:           229,
				ExpenseOccuredAt: time.Unix(1761249149, 0),
				Description:      "new altoids",
			},
		},
		{
			name: "valid-second-full-record",
			inputRecord: &expenses.Expense{
				Amount:           3278900,
				ExpenseOccuredAt: time.Unix(1761242999, 0),
				Description:      "a brand new car",
			},
			expectError: false,
			wantError:   nil,
			wantRecord: &expenses.Expense{
				ID:               7,
				Amount:           3278900,
				ExpenseOccuredAt: time.Unix(1761242999, 0),
				Description:      "a brand new car",
			},
		},
		{
			name:        "invalid-nil-record",
			inputRecord: nil,
			expectError: true,
			wantError:   sqlite.ErrNilPointer,
			wantRecord:  nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := sqlite.NewSqliteRepository(db)

			// defer teardown
			defer func() {
				err := db.Close()
				if err != nil {
					t.Errorf("unable to close connection to in-memory sqlite database: %v", err)
				}
			}()

			// call the function
			gotRecord, gotErr := repo.Create(context.Background(), testCase.inputRecord)

			// checking if we expect an error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("Create() got error: '%v', expected error: '%v'", gotErr, testCase.wantError)
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
