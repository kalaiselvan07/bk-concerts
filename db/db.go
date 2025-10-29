package db

import (
	"database/sql"
	"fmt"

	"bk-concerts/logger" // ⬅️ Assuming this import path

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// InitDB opens the database connection and assigns it to the global DB variable.
func InitDB(connStr string) error {
	var err error

	logger.Log.Info("[db] Attempting to open database connection...")
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[db] Error opening database: %v", err))
		return fmt.Errorf("error opening database: %w", err)
	}

	// Verify the connection is working
	logger.Log.Info("[db] Pinging database to verify connection...")
	if err = DB.Ping(); err != nil {
		logger.Log.Error(fmt.Sprintf("[db] Failed to ping database: %v", err))
		return fmt.Errorf("error pinging database: %w", err)
	}

	logger.Log.Info("[db] Successfully connected to PostgreSQL!")
	// fmt.Println("Successfully connected to PostgreSQL!") // Removed redundant fmt.Println in favor of logger
	return nil
}
