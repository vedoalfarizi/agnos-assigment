package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName           string
	AppEnv            string
	ServerPort        int
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	JWTSecret         string
	JWTExpirationDays int
	LogLevel          string
	LogFormat         string
	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppName:           getEnv("APP_NAME", "hospital-api"),
		AppEnv:            getEnv("APP_ENV", "development"),
		ServerPort:        getEnvInt("APP_PORT", 8080),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", "hospital_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTExpirationDays: getEnvInt("JWT_EXPIRATION_DAYS", 30),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),
		RateLimitEnabled:  getEnvBool("RATE_LIMIT_ENABLED", true),
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
	}

	lifetime := getEnv("DB_CONN_MAX_LIFETIME", "5m")
	duration, err := time.ParseDuration(lifetime)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}
	cfg.DBConnMaxLifetime = duration

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
