// Package routes handles the routes and what controller (service) method to utilize.
package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/handler"
)

func logger(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.Method, r.RequestURI, time.Since(start), r.RemoteAddr)
	})
}

func SetupRoutes(service expenses.Service) (*http.ServeMux, error) {
	m := http.NewServeMux()
	h := handler.NewExpanseHandler(service)

	// register the routes and return
	m.HandleFunc("GET /expenses", logger(h.GetAllExpenses))
	m.HandleFunc("GET /expenses/{id}", logger(h.GetExpenseByID))
	m.HandleFunc("POST /expenses", logger(h.CreateExpense))
	m.HandleFunc("PUT /expenses", logger(h.UpdateExpense))
	m.HandleFunc("DELETE /expenses/{id}", logger(h.DeleteExpense))

	return m, nil
}
