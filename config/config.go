package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/zechao/faceit-user-svc/postgres"
)

type Config struct {
	APPEnv               string
	HTTPHost             string
	HTTPPort             string
	JWTSecret            string
	JWTExpirationSecoond int
	postgres.Config
}

var ENVs = initConfig()

// by default load .env file if APP_ENV is local
// otherwise load environment variables from the system
func initConfig() Config {
	appEnv := "local"
	if e := os.Getenv("APP_ENV"); e != "" {
		appEnv = e
	}
	if appEnv == "local" {
		// it wont load env variable if is already set in the system
		godotenv.Load()
	}

	debug, _ := strconv.ParseBool(getEnv("DEBUG_MODE", "false"))
	return Config{
		APPEnv:   appEnv,
		HTTPHost: getEnv("HTTP_HOST", "localhost"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		Config: postgres.Config{
			DBUser:     getEnv("DB_USER", "user"),
			DBName:     getEnv("DB_NAME", "user"),
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPassword: getEnv("DB_PASSWORD", "ecom"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
			DebugMode:  debug,
		},
	}

}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(value)
		if err != nil {
			log.Panicf("invalid int value for key %s", key)
		}
		return v
	}
	return fallback
}
