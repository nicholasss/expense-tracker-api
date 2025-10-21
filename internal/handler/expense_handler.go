package handler

import (
	"net/http"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

type ExpenseHandler struct {
	Service *expenses.Service
}

func NewExpanseHandler(service *expenses.Service) *ExpenseHandler {
	return &ExpenseHandler{Service: service}
}

func (h *ExpenseHandler) GetAllExpenses(w http.ResponseWriter, r *http.Request) {
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
}
