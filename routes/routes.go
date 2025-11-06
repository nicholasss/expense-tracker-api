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

func SetupRoutes(service *expenses.Service) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	han := handler.NewExpanseHandler(service)

	// register the routes and return
	mux.HandleFunc("GET /expenses", logger(han.GetAllExpenses))
	mux.HandleFunc("POST /expenses", logger(han.CreateExpense))

	return mux, nil
}
