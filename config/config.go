// Package config
package config

import (
	"os"

	"github.com/joho/godotenv"
)

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
	// sqlite
	DBString string
	DBDriver string
	// mongodb
	MongoDBURI string
}

// LoadConfig will load given file path and setup the config
func LoadConfig(filePath string) (*Config, error) {
	err := godotenv.Load(filePath)
	if err != nil {
		return nil, err
	}

	localAddress := os.Getenv("LOCAL_ADDRESS")
	localPort := os.Getenv("LOCAL_PORT")
	dbPath := os.Getenv("DB_PATH") // aka, database string
	dbDriver := os.Getenv("GOOSE_DRIVER")
	mongoDBURI := os.Getenv("MONGODB_URI")

	if localAddress == "" || localPort == "" || dbPath == "" || dbDriver == "" || mongoDBURI == "" {
		return nil, &MissingVariableError{}
	}

	conf := Config{
		// network
		LocalAddress: localAddress,
		LocalPort:    localPort,
		Address:      localAddress + ":" + localPort,

		// database
		DBString:   dbPath,
		DBDriver:   dbDriver,
		MongoDBURI: mongoDBURI,
	}

	return &conf, nil
}
