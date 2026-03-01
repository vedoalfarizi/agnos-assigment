package main

import (
	"fmt"
	"os"

	"github.com/vedoalfarizi/hospital-api/internal/config"
	"github.com/vedoalfarizi/hospital-api/internal/database/postgre"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)

	// establish database connection
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		// fall back to constructing from individual fields if needed
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	}
	if _, err := postgre.Connect(dsn); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	r := router.New(log, cfg)
	addr := fmt.Sprintf(":%d", cfg.ServerPort)

	log.Infof("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Errorf("Server error: %v", err)
		os.Exit(1)
	}
}
