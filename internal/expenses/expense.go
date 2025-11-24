package expenses

import "time"

// Expense is used for all expense types, except summaries
//
// ID & RecordCreatedAt is set in the repository layer
type Expense struct {
	ID               int       // id of the expense for db
	Amount           int64     // cents total
	ExpenseOccuredAt time.Time // when it happened
	RecordCreatedAt  time.Time // when the record was created
	Description      string    // what the transaction is
}
