package db

import (
	"database/sql"
	"fmt"
	"net/url"

	"supra/logger"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// InitDB opens the database connection and assigns it to the global DB variable.
func InitDB(connStr string) error {
	var err error

	logger.Log.Info("[db] Attempting to open database connection...")

	// ✅ Properly parse and modify DSN for lib/pq (safe & supported)
	u, err := url.Parse(connStr)
	if err != nil {
		return fmt.Errorf("invalid DB connection string: %w", err)
	}
	q := u.Query()
	q.Set("binary_parameters", "no") // <-- this is valid for lib/pq
	u.RawQuery = q.Encode()
	finalConnStr := u.String()

	// ✅ Open DB using the official lib/pq driver
	DB, err = sql.Open("postgres", finalConnStr)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[db] Error opening database: %v", err))
		return fmt.Errorf("error opening database: %w", err)
	}

	// ✅ Connection pool tuning
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	// ✅ Verify connection
	logger.Log.Info("[db] Pinging database to verify connection...")
	if err = DB.Ping(); err != nil {
		logger.Log.Error(fmt.Sprintf("[db] Failed to ping database: %v", err))
		return fmt.Errorf("error pinging database: %w", err)
	}

	logger.Log.Info("[db] Successfully connected to PostgreSQL!")
	return nil
}
