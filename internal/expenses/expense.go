package expenses

import "time"

type Expense struct {
	ID               int       // id of the expense for db
	Amount           int       // cents total
	ExpenseOccuredAt time.Time // when it happened
	RecordCreatedAt  time.Time // when the record was created
	Description      string    // what the transaction is
}
