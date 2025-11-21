// Package expenses implements the "business logic" for handling expenses and the repository interface for interacting with databases
package expenses

import (
	"context"
	"errors"
)

// ErrNilPointer is returned when a nil pointer dereference is avoided
var ErrNilPointer = errors.New("input pointer cannot be nil")

// ErrNoRowsDeleted is returned when a delete query does not affect any rows
var ErrNoRowsDeleted = errors.New("no rows were deleted")

// ErrNoRowsUpdated is returned when an update query does not affect any rows
var ErrNoRowsUpdated = errors.New("no rows were updated")

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
