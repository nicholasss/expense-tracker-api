package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/sqlite"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request processing...\n")
	_, err := fmt.Fprintln(w, "hello")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	conf, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	exp := &expenses.Expense{
		Amount:           1499,
		ExpenseOccuredAt: time.Now(),
		Description:      "new coffee for moka pot maker",
	}

	repo := sqlite.NewSqliteRepository(conf.database)
	exp, err = repo.Create(context.Background(), exp)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n", exp)

	// setup new mux
	mux := http.NewServeMux()

	// root handler
	mux.HandleFunc("/", rootHandler)

	log.Printf("Starting server...\n")

	err = http.ListenAndServe(conf.toAddr(), mux)
	if err != nil {
		log.Fatal(err)
	}
}
