// Package routes
package routes

import (
	"net/http"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
)

func SetupRoutes(service *expenses.Service) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	return mux, nil
}
