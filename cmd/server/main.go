package main

import (
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nicholasss/expense-tracker-api/config"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/sqlite"
	"github.com/nicholasss/expense-tracker-api/routes"
)

const ConfigPath = ".env"

func main() {
	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	repository := sqlite.NewSqliteRepository(cfg.DB)
	service := expenses.NewService(repository)

	mux, err := routes.SetupRoutes(service)
	if err != nil {
		log.Fatalf("Failed to setup routes: %v", err)
	}

	log.Printf("Starting server at %s...\n", cfg.Address)
	err = http.ListenAndServe(cfg.Address, mux)
	if err != nil {
		log.Fatal(err)
	}
}
