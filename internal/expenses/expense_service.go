package expenses

import (
	"context"
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

// Service implements all of the underlying business logic.
// Things such as expenses being positive and not zero, etc.
type Service struct {
	repo Repository
}

// NewService utilizes the Repository interface defined in internal/repository.go
// This way, we never need to worry about the underlying database
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) NewExpense(ctx context.Context, occuredAt time.Time, description string, amount int64) (*Expense, error) {
	// check description
	if description == "" {
		return nil, ErrInvalidDescription
	}

	// check amount
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// able to be unix time
	unixEpoch := time.Unix(0, 0)
	if !occuredAt.After(unixEpoch) {
		return nil, ErrInvalidOccuredAtTime
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

func (s *Service) GetAllExpenses(ctx context.Context) ([]*Expense, error) {
	exps, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return exps, nil
}

func (s *Service) GetExpenseByID(ctx context.Context, id int) (*Expense, error) {
	if id <= 0 {
		return nil, fmt.Errorf("id needs to be greater than 0")
	}

	exp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return exp, nil
}

func (s *Service) UpdateExpense(ctx context.Context, id int, occuredAt time.Time, description string, amount int64) error {
	if id <= 0 {
		return fmt.Errorf("id needs to be greater than 0")
	}

	// check description
	if description == "" {
		return fmt.Errorf("empty expense description")
	}

	// check amount
	if amount <= 0 {
		return fmt.Errorf("expense amount needs to be greater than 0")
	}

	// able to be unix time
	nineteenSeventy := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if occuredAt.Before(nineteenSeventy) {
		return fmt.Errorf("expense date needs to be after 1970")
	}

	exp := &Expense{
		Amount:           amount,
		ExpenseOccuredAt: occuredAt,
		Description:      description,
	}

	if err := s.repo.Update(ctx, exp); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteExpense(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("id needs to be greater than 0")
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
	monthStr, yearStr, found := strings.Cut(str, "-")
	if !found {
		return time.Time{}, fmt.Errorf("could not parse custom month: %q", str)
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, err
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return time.Time{}, err
	}

	customMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return customMonth, nil
}

// isWrongYear is utilized within slices.DeleteFunc().
// It will return true if the two times are not the same year (year only),
// false if it is the same year.
func isWrongYear(timeA, timeB time.Time) bool {
	return timeA.Year() != timeB.Year()
}

func makeCustomYear(str string) (time.Time, error) {
	year, err := strconv.Atoi(str)
	if err != nil {
		return time.Time{}, err
	}

	customYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	return customYear, nil
}

func (s *Service) SummarizeExpenses(ctx context.Context, kind SummaryTimeRange, modifier string) (*ExpenseSummary, error) {
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
		log.Println("custom range not implemented yet")
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
