package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	Environment         = ""
	SqlConnectionString = ""
	JwtSecret           = ""
	ExpirationAt        = 0
)

const configDir = "configs"

func SetupEnvironments() {
	if err := Load(); err != nil {
		log.Fatal(err)
	}
}

func Load() error {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		return errors.New("environment variable ENVIRONMENT is required")
	}

	v := viper.New()
	configPath, err := resolveConfigPath(environment)
	if err != nil {
		return err
	}
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file %q: %w", configPath, err)
	}

	Environment = v.GetString("environment")
	SqlConnectionString = v.GetString("mssql.connectionString")
	JwtSecret = v.GetString("security.jwtSecret")
	ExpirationAt = v.GetInt("security.expirationAt")

	return nil
}

func resolveConfigPath(environment string) (string, error) {
	fileName := fmt.Sprintf("config.%s.yaml", environment)
	workdir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		candidate := filepath.Join(workdir, configDir, fileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(workdir)
		if parent == workdir {
			return "", fmt.Errorf("config file %q not found under %q", fileName, configDir)
		}

		workdir = parent
	}
}
