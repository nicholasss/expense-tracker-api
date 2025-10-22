// Package config
package config

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"
)

const database string = "expense-tracker.db"

type MissingVariableError struct{}

func (e *MissingVariableError) Error() string {
	return "missing required environmental variable(s)"
}

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
	err := godotenv.Load(filePath)
	if err != nil {
		return nil, err
	}

	localAddress := os.Getenv("LOCAL_ADDRESS")
	localPort := os.Getenv("LOCAL_PORT")
	dbPath := os.Getenv("DB_PATH")
	dbDriver := os.Getenv("GOOSE_DRIVER")

	if localAddress == "" || localPort == "" || dbPath == "" || dbDriver == "" {
		return nil, &MissingVariableError{}
	}

	db, err := sql.Open(dbDriver, database)
	if err != nil {
		return nil, err
	}

	conf := Config{
		// network
		LocalAddress: localAddress,
		LocalPort:    localPort,
		Address:      localAddress + ":" + localPort,

		// database
		DBPath:   dbPath,
		DBDriver: dbDriver,
		DB:       db,
	}

	return &conf, nil
}
