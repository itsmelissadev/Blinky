package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPass           string
	DBName           string
	DBConnString     string
	PublicAPIHost    string
	PublicAPIPort    string
	AdminPanelHost   string
	AdminPanelPort   string
	PostgresPath     string
	PostgresDataPath string
	IsEnvExist       bool
	AdminSSHEnabled  bool
	PublicSSHEnabled bool
	SSHPort          string
	SSHUser          string
	SSHPassword      string
}

func LoadConfig() *Config {
	isEnvExist := true
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		isEnvExist = false
	}

	_ = godotenv.Load()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "postgres")
	dbName := getEnv("DB_NAME", "blinky_db")

	return &Config{
		DBHost:           dbHost,
		DBPort:           dbPort,
		DBUser:           dbUser,
		DBPass:           dbPass,
		DBName:           dbName,
		DBConnString:     fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName),
		PublicAPIHost:    getEnv("PUBLIC_API_HOST", "0.0.0.0"),
		PublicAPIPort:    getEnv("PUBLIC_API_PORT", "8090"),
		AdminPanelHost:   getEnv("ADMIN_PANEL_HOST", "0.0.0.0"),
		AdminPanelPort:   getEnv("ADMIN_PANEL_PORT", "8080"),
		PostgresPath:     getEnv("POSTGRESQL_FOLDER_PATH", ""),
		PostgresDataPath: getEnv("POSTGRESQL_DATA_PATH", ""),
		IsEnvExist:       isEnvExist,

		AdminSSHEnabled:  getEnv("ADMIN_SSH_ENABLED", "false") == "true",
		PublicSSHEnabled: getEnv("PUBLIC_SSH_ENABLED", "false") == "true",
		SSHPort:          getEnv("SSH_PORT", ""),
		SSHUser:          getEnv("SSH_USER", ""),
		SSHPassword:      getEnv("SSH_PASS", ""),
	}
}

func UpdateEnvVariables(updates map[string]string) error {
	env, _ := godotenv.Read()
	if env == nil {
		env = make(map[string]string)
	}

	for k, v := range updates {
		env[k] = v
		_ = os.Setenv(k, v)
	}

	return godotenv.Write(env, ".env")
}

func SaveEnv(data map[string]string) error {
	return godotenv.Write(data, ".env")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
