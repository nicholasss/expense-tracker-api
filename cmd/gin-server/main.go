package main

import (
	"log"

	_ "go.mongodb.org/mongo-driver/v2/mongo"
	// _ "github.com/mattn/go-sqlite3"

	"github.com/nicholasss/expense-tracker-api/config"
	"github.com/nicholasss/expense-tracker-api/internal/expenses"
	"github.com/nicholasss/expense-tracker-api/internal/mongodb"
	"github.com/nicholasss/expense-tracker-api/routes"
)

const ConfigPath = ".env"

func main() {
	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// repository := sqlite.NewSqliteRepository(cfg.DB)
	repository, err := mongodb.NewMongoDBRespository(cfg.MongoDBURI)
	if err != nil {
		log.Fatal(err)
	}

	service := expenses.NewService(repository)

	router := routes.SetupGinRoutes(service)
	log.Printf("Starting server at %s...\n", cfg.Address)

	err = router.Run(cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
}
