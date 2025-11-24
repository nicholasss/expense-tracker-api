package main

import (
	"errors"
	"log"

	"github.com/nicholasss/expense-tracker-api/config"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/mongodb"
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

	repository, err := mongodb.NewMongoDBRespository(cfg.MongoDBURI)
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
