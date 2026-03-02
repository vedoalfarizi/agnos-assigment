package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vedoalfarizi/hospital-api/internal/config"
	"github.com/vedoalfarizi/hospital-api/internal/database/postgre"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
	"github.com/vedoalfarizi/hospital-api/internal/router"
)

// Bootstrap holds all initialized application components
type Bootstrap struct {
	Config *config.Config
	Logger *logger.Logger
	Router *gin.Engine
}

// Run initializes all application dependencies in the correct order
func (b *Bootstrap) Run() error {
	// Step 1: Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	b.Config = cfg

	// Step 2: Initialize logger
	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	b.Logger = log

	// Step 3: Establish database connection
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		// fall back to constructing from individual fields if needed
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	}
	if _, err := postgre.Connect(dsn); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Step 4: Setup router with all dependencies
	db := postgre.GetDB()
	r := router.New(log, cfg, db)
	b.Router = r

	return nil
}

// Shutdown gracefully closes application resources
func (b *Bootstrap) Shutdown(ctx context.Context) error {
	if b.Logger != nil {
		b.Logger.Info("Shutting down application...")
	}

	// Close database connection with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if db := postgre.GetDB(); db != nil {
		if err := db.Close(); err != nil {
			if b.Logger != nil {
				b.Logger.Errorf("Error closing database connection: %v", err)
			}
			return fmt.Errorf("failed to close database: %w", err)
		}
		if b.Logger != nil {
			b.Logger.Info("Database connection closed")
		}
	}

	// Give pending requests a chance to complete
	select {
	case <-shutdownCtx.Done():
		if b.Logger != nil {
			b.Logger.Warn("Shutdown timeout exceeded")
		}
		return shutdownCtx.Err()
	default:
	}

	if b.Logger != nil {
		b.Logger.Info("Shutdown complete")
	}
	return nil
}

// GetServerAddr returns the formatted server address
func (b *Bootstrap) GetServerAddr() string {
	if b.Config == nil {
		return ":8080"
	}
	return fmt.Sprintf(":%d", b.Config.ServerPort)
}
