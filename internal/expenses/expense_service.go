package expenses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"
)

type SummaryTimeRange int

const (
	AllExpenses SummaryTimeRange = iota
	ThisMonth
	CustomMonth
	ThisYear
	CustomYear
	CustomYearMonthRange
)

// These errors are used in the validation step of NewExpense()
var (
	ErrInvalidDescription   = fmt.Errorf("expense description cannot be empty")
	ErrInvalidAmount        = fmt.Errorf("expense amount needs to be greater than 0")
	ErrInvalidOccuredAtTime = fmt.Errorf("expense date needs to be after 1970")
)

// ErrInvalidID is used with validation step of GetExpenseByID()
var ErrInvalidID = fmt.Errorf("id needs to be greater than 0")

// ErrUnusedID is used in the validation step of GetExpenseByID(),
// for record ID's that structurally valid (above 0) but do not have a valid record
var ErrUnusedID = fmt.Errorf("id used does not have a valid record")

// ErrInvalidTime is used for SummarizeExpenses() when an invalid range is provided
type ErrInvalidTime struct {
	ProvidedTime string
	WrappedError error
}

func (e *ErrInvalidTime) Error() string {
	if e.WrappedError != nil {
		return fmt.Sprintf("invalid time range of '%s' due to: '%s'", e.ProvidedTime, e.WrappedError)
	}
	return fmt.Sprintf("invalid time range of '%s'", e.ProvidedTime)
}

// checkDescription is to validate an expenses description
func checkDescription(description string) error {
	if description == "" {
		return ErrInvalidDescription
	}
	return nil
}

// checkAmount is to validate an expenses amount
func checkAmount(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// checkOccuredAt is to validate an expenses occuredAt time
func checkOccuredAt(occ time.Time) error {
	unixEpoch := time.Unix(0, 0)
	if !occ.After(unixEpoch) {
		return ErrInvalidOccuredAtTime
	}
	return nil
}

// ExpenseService implements all of the underlying business logic.
// Things such as expenses being positive and not zero, etc.
type ExpenseService struct {
	repo Repository
}

// NewService utilizes the Repository interface defined in internal/repository.go
// This way, we never need to worry about the underlying database
func NewService(repo Repository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) NewExpense(ctx context.Context, occuredAt time.Time, description string, amount int64) (*Expense, error) {
	// check description
	if err := checkDescription(description); err != nil {
		return nil, err
	}

	// check amount
	if err := checkAmount(amount); err != nil {
		return nil, err
	}

	// able to be unix time
	if err := checkOccuredAt(occuredAt); err != nil {
		return nil, err
	}

	exp := &Expense{
		Amount:           amount,
		ExpenseOccuredAt: occuredAt,
		Description:      description,
	}

	exp, err := s.repo.Create(ctx, exp)
	if err != nil {
		return nil, err
	}

	return exp, nil
}

func (s *ExpenseService) GetAllExpenses(ctx context.Context) ([]*Expense, error) {
	exps, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return exps, nil
}

func (s *ExpenseService) GetExpenseByID(ctx context.Context, id int) (*Expense, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	exp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUnusedID
		}
		return nil, err
	}

	return exp, nil
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, id int, occuredAt time.Time, description string, amount int64) error {
	if id <= 0 {
		return ErrInvalidID
	}

	// check description
	if err := checkDescription(description); err != nil {
		return err
	}

	// check amount
	if err := checkAmount(amount); err != nil {
		return err
	}

	// able to be unix time
	if err := checkOccuredAt(occuredAt); err != nil {
		return err
	}

	exp := &Expense{
		ID:               id,
		Amount:           amount,
		ExpenseOccuredAt: occuredAt,
		Description:      description,
	}

	if err := s.repo.Update(ctx, exp); err != nil {
		return err
	}

	return nil
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidID
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

// isWrongMonth is utilized within slices.DeleteFunc().
// It will return true if the two times are not the same month (year and month),
// false if it is the same month.
func isWrongMonth(timeA, timeB time.Time) bool {
	return timeA.Year() != timeB.Year() || timeA.Month() != timeB.Month()
}

func makeCustomMonth(str string) (time.Time, error) {
	yearStr, monthStr, found := strings.Cut(str, "-")
	if !found {
		return time.Time{}, &ErrInvalidTime{ProvidedTime: str, WrappedError: nil}
	}

	monthInt, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, &ErrInvalidTime{ProvidedTime: str, WrappedError: err}
	}
	yearInt, err := strconv.Atoi(yearStr)
	if err != nil {
		return time.Time{}, &ErrInvalidTime{ProvidedTime: str, WrappedError: err}
	}

	// perform explicit check for month before type casting to time.Month
	if monthInt < 1 || monthInt > 12 {
		return time.Time{}, &ErrInvalidTime{ProvidedTime: str}
	}

	customMonth := time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC)

	// ensure its valid unix time
	unixEpoch := time.Unix(0, 0)
	if customMonth.Before(unixEpoch) {
		return time.Time{}, &ErrInvalidTime{ProvidedTime: str}
	}

	return customMonth, nil
}

// isWrongYear is utilized within slices.DeleteFunc().
// It will return true if the two times are not the same year (year only),
// false if it is the same year.
func isWrongYear(timeA, timeB time.Time) bool {
	return timeA.Year() != timeB.Year()
}

func makeCustomYear(str string) (time.Time, error) {
	yearInt, err := strconv.Atoi(str)
	if err != nil {
		return time.Time{}, &ErrInvalidTime{ProvidedTime: str, WrappedError: err}
	}

	customYear := time.Date(yearInt, 1, 0, 0, 0, 0, 0, time.UTC)

	// ensure its valid unix time
	unixEpoch := time.Unix(0, 0)
	if customYear.Before(unixEpoch) {
		return time.Time{}, &ErrInvalidTime{ProvidedTime: str}
	}

	return customYear, nil
}

func (s *ExpenseService) SummarizeExpenses(ctx context.Context, kind SummaryTimeRange, modifier string) (*ExpenseSummary, error) {
	exps, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var summaryTimeRangeString string

	now := time.Now().UTC()
	// filter out what doesnt match
	switch kind {
	case AllExpenses:
		// implicit brake
	case ThisMonth:
		summaryTimeRangeString = "This Month"

		exps = slices.DeleteFunc(exps, func(exp *Expense) bool {
			return isWrongMonth(exp.ExpenseOccuredAt, now)
		})
	case CustomMonth:
		// i.e. '2024-01'
		customMonth, err := makeCustomMonth(modifier)
		if err != nil {
			return nil, err
		}

		summaryTimeRangeString = fmt.Sprintf("Custom Month: %s of %d", customMonth.Month(), customMonth.Year())

		exps = slices.DeleteFunc(exps, func(exp *Expense) bool {
			return isWrongMonth(exp.ExpenseOccuredAt, customMonth)
		})
	case ThisYear:
		summaryTimeRangeString = "This Year"

		exps = slices.DeleteFunc(exps, func(exp *Expense) bool {
			return isWrongMonth(exp.ExpenseOccuredAt, now)
		})
	case CustomYear:
		customYear, err := makeCustomYear(modifier)
		if err != nil {
			return nil, err
		}

		summaryTimeRangeString = fmt.Sprintf("Custom Year: %d", customYear.Year())

		exps = slices.DeleteFunc(exps, func(exp *Expense) bool {
			return isWrongYear(exp.ExpenseOccuredAt, customYear)
		})
	case CustomYearMonthRange:
		// TODO: implement CustomYearMonthRange
		// "2023-09,2024-09", comma seperating out range begin and range end
		log.Println("WARNING: custom range not implemented yet")
	}

	// add up expenses
	var expenseSum int64
	for _, exp := range exps {
		expenseSum += exp.Amount
	}

	expSum := &ExpenseSummary{
		SummaryTimeRange: summaryTimeRangeString,
		Total:            expenseSum,
	}

	return expSum, nil
}
