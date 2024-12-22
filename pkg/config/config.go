package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Host config
	HostPort string

	// Db Config
	DbUri  string
	DbName string

	// App config
	AppEnv   string
	IsDevEnv bool
}

func NewConfig() *Config {
	godotenv.Load()
	env := strings.ToLower(getEnvString("ENV", "production"))

	return &Config{
		HostPort: getEnvString("HOST_PORT", "8080"),
		DbUri:    getEnvString("DB_URI", "mongodb://localhost:27017"),
		DbName:   getEnvString("DB_NAME", "defectdetect"),
		AppEnv:   env,
		IsDevEnv: env == "dev",
	}
}

func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
