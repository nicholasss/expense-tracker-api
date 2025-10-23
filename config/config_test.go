package config_test

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nicholasss/expense-tracker-api/config"
)

// checkConfigEquality is used in testing to compare two different *config.Config structs
func checkConfigEquality(t *testing.T, got, want *config.Config) {
	t.Helper()

	// network
	if got.LocalAddress != want.LocalAddress {
		t.Errorf("conf.LocalAddress does not match. got: '%v', want: '%v'", got.LocalAddress, want.LocalAddress)
	}
	if got.LocalPort != want.LocalPort {
		t.Errorf("conf.LocalPort does not match. got: %v, want: '%v'", got.LocalPort, want.LocalPort)
	}
	if got.Address != want.Address {
		t.Errorf("conf.Address does not match. got: '%v', want: '%v'", got.Address, want.Address)
	}

	// database
	if got.DBPath != want.DBPath {
		t.Errorf("conf.DBPath does not match. got: '%v', want: '%v'", got.DBPath, want.DBPath)
	}
	if got.DBDriver != want.DBDriver {
		t.Errorf("conf.DBDriver does not match. got: '%v', want: '%v'", got.DBDriver, want.DBDriver)
	}
}

func unsetEnvVars(t *testing.T, keyList []string) {
	t.Helper()

	for _, key := range keyList {
		err := os.Unsetenv(key)
		if err != nil {
			t.Errorf("unable to unset key %q", key)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	envVarKeys := []string{
		"LOCAL_ADDRESS",
		"LOCAL_PORT",
		"DB_PATH",
		"GOOSE_DRIVER",
		"GOOSE_DBSTRING",
	}

	testTable := []struct {
		name        string
		inputConfig string
		expectError bool
		wantError   error
		wantConfig  *config.Config
	}{
		{
			name: "valid-config-load",
			inputConfig: `# server vars
      export LOCAL_ADDRESS="localhost"
      export LOCAL_PORT="8080"
      export DB_PATH="./expense-tracker.db"

      # Goose vars
      export GOOSE_DRIVER="sqlite3"
      export GOOSE_DBSTRING="./../../expense-tracker.db"`,
			expectError: false,
			wantError:   nil,
			wantConfig: &config.Config{
				LocalAddress: "localhost",
				LocalPort:    "8080",
				Address:      "localhost:8080",
				DBPath:       "./expense-tracker.db",
				DBDriver:     "sqlite3",
			},
		},
		{
			name: "valid-config-load",
			inputConfig: `# server vars
      export LOCAL_ADDRESS="localhost"
      export LOCAL_PORT="8080"
      export DB_PATH="./expense-tracker.db"

      # Goose vars
      export GOOSE_DRIVER="sqlite3"`,
			expectError: false,
			wantError:   nil,
			wantConfig: &config.Config{
				LocalAddress: "localhost",
				LocalPort:    "8080",
				Address:      "localhost:8080",
				DBPath:       "./expense-tracker.db",
				DBDriver:     "sqlite3",
			},
		},
		{
			name:        "invalid-empty-config-load",
			inputConfig: ``,
			expectError: true,
			wantError:   &config.MissingVariableError{},
			wantConfig:  nil,
		},
		{
			name: "invalid-partial-config-load",
			inputConfig: `# server vars
      export LOCAL_ADDRESS="localhost"
      export LOCAL_PORT="8080"
      export DB_PATH="./expense-tracker.db"`,
			expectError: true,
			wantError:   &config.MissingVariableError{},
			wantConfig:  nil,
		},
		{
			name: "invalid-partial-config-load",
			inputConfig: `# Goose vars
      export GOOSE_DRIVER="sqlite3"
      export GOOSE_DBSTRING="./../../expense-tracker.db"`,
			expectError: true,
			wantError:   &config.MissingVariableError{},
			wantConfig:  nil,
		},
		{
			name: "invalid-missing-one-config-load",
			inputConfig: `# server vars
      export LOCAL_PORT="8080"
      export DB_PATH="./expense-tracker.db"

      # Goose vars
      export GOOSE_DRIVER="sqlite3"
      export GOOSE_DBSTRING="./../../expense-tracker.db"`,
			expectError: true,
			wantError:   &config.MissingVariableError{},
			wantConfig:  nil,
		},
		{
			name: "invalid-missing-one-config-load",
			inputConfig: `# server vars
      export LOCAL_ADDRESS="localhost"
      export DB_PATH="./expense-tracker.db"

      # Goose vars
      export GOOSE_DRIVER="sqlite3"
      export GOOSE_DBSTRING="./../../expense-tracker.db"`,
			expectError: true,
			wantError:   &config.MissingVariableError{},
			wantConfig:  nil,
		},
		{
			name: "invalid-missing-one-config-load",
			inputConfig: `# server vars
      export LOCAL_ADDRESS="localhost"
      export LOCAL_PORT="8080"

      # Goose vars
      export GOOSE_DRIVER="sqlite3"
      export GOOSE_DBSTRING="./../../expense-tracker.db"`,
			expectError: true,
			wantError:   &config.MissingVariableError{},
			wantConfig:  nil,
		},
	}

	// actual tests here
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// call an unset env func
			unsetEnvVars(t, envVarKeys)

			// creating tmp .env file
			tmpFile, err := os.CreateTemp("/tmp", "*.env")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			// removing tmp .env file
			defer func() {
				err = os.Remove(tmpFile.Name())
				if err != nil {
					t.Errorf("unable to remove temp file: %v", tmpFile.Name())
				}
			}()

			// write to temp file
			err = os.WriteFile(tmpFile.Name(), []byte(testCase.inputConfig), 0o644)
			if err != nil {
				t.Fatalf("failed to write to temporary file: %v", err)
			}

			// call the function
			gotConfig, gotErr := config.LoadConfig(tmpFile.Name())

			if (gotErr != nil) != testCase.expectError {
				t.Errorf("LoadConfig(%q) with config: %v, got error: '%v', expected error: '%v'", tmpFile.Name(), testCase.inputConfig, gotErr, testCase.wantError)
			}

			// safer comparison
			if !testCase.expectError && gotConfig != nil {
				checkConfigEquality(t, gotConfig, testCase.wantConfig)
			}
		})
	}
}
