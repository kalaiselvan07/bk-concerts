package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// InitDB opens the database connection and assigns it to the global DB variable.
func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Verify the connection is working
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")
	return nil
}
