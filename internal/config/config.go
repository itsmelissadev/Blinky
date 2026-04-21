package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnString     string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPass           string
	DBName           string
	PublicAPIHost    string
	PublicAPIPort    string
	AdminPanelHost   string
	AdminPanelPort   string
	PostgresPath     string
	PostgresDataPath string
	IsEnvExist       bool
}

func LoadConfig() *Config {
	isEnvExist := true
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		isEnvExist = false
	}

	if isEnvExist {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: Could not load .env file: %v\n", err)
		}
	}

	getEnv := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

	dbHost := getEnv("POSTGRESQL_DB_HOST", "")
	dbPort := getEnv("POSTGRESQL_DB_PORT", "")
	dbUser := getEnv("POSTGRESQL_DB_USER", "")
	dbPass := getEnv("POSTGRESQL_DB_PASSWORD", "")
	dbName := getEnv("POSTGRESQL_DB_NAME", "")
	postgresPath := getEnv("POSTGRESQL_FOLDER_PATH", "")
	postgresDataPath := getEnv("POSTGRESQL_DATA_PATH", "")

	return &Config{
		DBHost: dbHost,
		DBPort: dbPort,
		DBUser: dbUser,
		DBPass: dbPass,
		DBName: dbName,
		DBConnString: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPass, dbHost, dbPort, dbName,
		),
		PublicAPIHost:    getEnv("PUBLIC_API_HOST", "localhost"),
		PublicAPIPort:    getEnv("PUBLIC_API_PORT", "8090"),
		AdminPanelHost:   getEnv("ADMIN_PANEL_HOST", "localhost"),
		AdminPanelPort:   getEnv("ADMIN_PANEL_PORT", "8080"),
		PostgresPath:     postgresPath,
		PostgresDataPath: postgresDataPath,
		IsEnvExist:       isEnvExist,
	}
}

func SaveEnv(data map[string]string) error {
	var content strings.Builder
	for k, v := range data {
		content.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	return os.WriteFile(".env", []byte(content.String()), 0644)
}
