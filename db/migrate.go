package db

import (
	"fmt"

	"supra/logger"
)

const createSeatTableSQL = `
CREATE TABLE IF NOT EXISTS seat (
    seat_id UUID PRIMARY KEY,
    seat_type TEXT NOT NULL,
    price_gel REAL NOT NULL,
    price_inr REAL NOT NULL,
    available INTEGER NOT NULL,
    notes TEXT
);`

const createConcertTableSQL = `
CREATE TABLE IF NOT EXISTS concert (
    concert_id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    venue TEXT NOT NULL,
    timing TEXT NOT NULL,
    seat_ids JSONB,  -- Using JSONB is recommended for storing JSON in Postgres
    payment_ids JSONB,
    payment_details_ids JSONB,
    description TEXT
);`

const createParticipantTableSQL = `
CREATE TABLE IF NOT EXISTS participant (
    user_id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    wa_num TEXT NOT NULL,
    email TEXT,
    attended BOOLEAN NOT NULL
);`

const createPaymentTableSQL = `
CREATE TABLE IF NOT EXISTS payment (
    payment_id UUID PRIMARY KEY,
    payment_type TEXT NOT NULL,
    details TEXT NOT NULL,
    notes TEXT
);`

const createBookingTableSQL = `
CREATE TABLE IF NOT EXISTS booking (
    booking_id UUID PRIMARY KEY,
    booking_email TEXT NOT NULL,
    booking_status TEXT NOT NULL,
    receipt_image BYTEA,
    seat_quantity INTEGER NOT NULL,
    seat_id TEXT NOT NULL,
    concert_id TEXT REFERENCES concert(concert_id),
    seat_type TEXT NOT NULL,
    total_amount REAL NOT NULL,
    participant_ids JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE
);`

const createUserTableSQL = `
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash BYTEA,              -- Stores bcrypt hash for admins
    role TEXT NOT NULL DEFAULT 'user', -- 'admin' or 'user'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);`

const createOTPTableSQL = `
CREATE TABLE IF NOT EXISTS otp_codes (
    code TEXT PRIMARY KEY,
    user_email TEXT UNIQUE NOT NULL REFERENCES users(email), 
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);`

// RunMigrations executes all necessary database structure changes.
func RunMigrations() error {
	if DB == nil {
		logger.Log.Error("[db] Cannot run migrations: Database connection is nil.")
		return fmt.Errorf("database connection is nil, call InitDB first")
	}

	// Define all migration steps in order, including their names for clear logs
	migrationSteps := []struct {
		Name string
		SQL  string
	}{
		{Name: "Users", SQL: createUserTableSQL},
		{Name: "OTP Codes", SQL: createOTPTableSQL},
		{Name: "Seats", SQL: createSeatTableSQL},
		{Name: "Concerts", SQL: createConcertTableSQL},
		{Name: "Participants", SQL: createParticipantTableSQL},
		{Name: "Payments", SQL: createPaymentTableSQL},
		{Name: "Bookings", SQL: createBookingTableSQL},
	}

	logger.Log.Info("[db] Starting database migrations...")

	for _, step := range migrationSteps {
		logger.Log.Info(fmt.Sprintf("[db] Running migration for table: %s", step.Name))

		// Execute the SQL, ensuring we're using the global DB variable defined in db.go
		if _, err := DB.Exec(step.SQL); err != nil {
			logger.Log.Error(fmt.Sprintf("[db] Failed migration for %s: %v", step.Name, err))
			// The return error strings are slightly inconsistent (e.g., "error crunning concert table migration"),
			// so I'm using a consistent format here.
			return fmt.Errorf("error running %s table migration: %w", step.Name, err)
		}
		logger.Log.Info(fmt.Sprintf("[db] Successfully migrated table: %s", step.Name))
	}

	logger.Log.Info("[db] All migrations completed successfully.")
	// fmt.Println("Migrations completed successfully.") // Removed redundant fmt.Println
	return nil
}
