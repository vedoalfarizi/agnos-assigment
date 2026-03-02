package postgre

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

// Connect opens a connection to the database using the given DSN and stores the
// resulting *sqlx.DB so it can be retrieved later with GetDB. The connection
// parameters are configured with sensible defaults that can be tuned if
// necessary. This function should be called during application startup.
func Connect(dsn string) (*sqlx.DB, error) {
	d, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// basic pooling settings
	d.SetMaxOpenConns(25)
	d.SetMaxIdleConns(5)
	d.SetConnMaxLifetime(5 * time.Minute)

	if err := d.Ping(); err != nil {
		return nil, err
	}

	db = d
	return db, nil
}

// GetDB returns the global *sqlx.DB instance previously created with Connect.
// Callers must ensure Connect has been invoked first.
func GetDB() *sqlx.DB {
	return db
}
