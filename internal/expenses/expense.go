package expenses

import "time"

// Expense is used for all expense types, except summaries
//
// ID & RecordCreatedAt is set in the repository layer
type Expense struct {
	ID               int       `json:"id"`          // id of the expense for db
	Amount           int64     `json:"amount"`      // cents total
	ExpenseOccuredAt time.Time `json:"occuredAt"`   // when it happened
	RecordCreatedAt  time.Time `json:"createdAt"`   // when the record was created
	Description      string    `json:"description"` // what the transaction is
}

// ExpenseSummary is used when the a summary is requested
type ExpenseSummary struct {
	SummaryTimeRange string `json:"summaryTimeRange"` // what time range
	Total            int64  `json:"total"`            // cents total
}
