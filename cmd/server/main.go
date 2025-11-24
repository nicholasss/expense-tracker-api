package main

import (
	"errors"
	"log"

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
		if errors.Is(err, &config.MissingVariableError{}) {
			log.Fatal("missing variable in .env")
		}

		log.Fatalf("Failed to load config: %v", err)
	}

	repository, err := sqlite.NewSqliteRepository(cfg.DBDriver, cfg.DBString)
	if err != nil {
		log.Fatalf("Failed to load SQLite3 database: %v", err)
	}

	service := expenses.NewService(repository)

	ginEngine := routes.SetupRoutes(service)
	log.Printf("Starting server at %s...\n", cfg.Address)

	err = ginEngine.Run(cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
}
