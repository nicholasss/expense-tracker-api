// Package routes
package routes

import (
	"net/http"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/handler"
)

func SetupRoutes(service *expenses.Service) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	expenseHandler := handler.NewExpanseHandler(service)

	// register the routes and return
	mux.HandleFunc("GET /expenses", expenseHandler.GetAllExpenses)

	return mux, nil
}
