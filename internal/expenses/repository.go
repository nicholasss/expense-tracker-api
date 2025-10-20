// Package expenses implements the "business logic" for handling expenses and the repository interface for interacting with databases
package expenses

import "context"

type Repository interface {
	// get one expense record by ID
	GetByID(ctx context.Context, id int) (*Expense, error)

	// get all expenses
	GetAll(ctx context.Context) ([]*Expense, error)

	// create a new expense
	Create(ctx context.Context, exp *Expense) (*Expense, error)

	// update an existing expense
	Update(ctx context.Context, exp *Expense) error

	// delete an exisiting expense
	Delete(ctx context.Context, id int) error
}
