// Package config
package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const database string = "expense-tracker.db"

type Config struct {
	// Network config
	LocalAddress string
	LocalPort    string
	// Hosting address, i.e. 10.0.0.1:8080
	Address string

	// Database config
	DBPath   string
	DBDriver string
	DB       *sql.DB
}

// LoadConfig will load given file path and setup the config
func LoadConfig(filePath string) (*Config, error) {
	if filePath != ".env" {
		log.Println("only supports .env loading currently, json coming soon.")
		return nil, nil
	}

	err := godotenv.Load(filePath)
	if err != nil {
		return nil, err
	}

	conf := Config{
		LocalAddress: os.Getenv("LOCAL_ADDRESS"),
		LocalPort:    os.Getenv("LOCAL_PORT"),
		DBPath:       os.Getenv("DB_PATH"),
		DBDriver:     os.Getenv("GOOSE_DRIVER"),
	}

	db, err := sql.Open(conf.DBDriver, database)
	if err != nil {
		return nil, err
	}

	// final configuation
	conf.DB = db
	conf.Address = conf.makeAddr()

	return &conf, nil
}

// Address returns the hosting address of the server
func (conf *Config) makeAddr() string {
	return conf.LocalAddress + ":" + conf.LocalPort
}
