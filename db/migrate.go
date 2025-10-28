// db/migrate.go
package db

import "fmt"

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
    payment_details_id TEXT NOT NULL,
    receipt_image BYTEA,
    seat_quantity INTEGER NOT NULL,
    seat_id TEXT NOT NULL,
    seat_type TEXT NOT NULL,
    total_amount REAL NOT NULL,
    participant_ids JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);`

// RunMigrations executes all necessary database structure changes.
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil, call InitDB first")
	}

	// Run Seat Migration
	if _, err := DB.Exec(createSeatTableSQL); err != nil {
		return fmt.Errorf("error running seat table migration: %w", err)
	}

	// Run Concert Migration
	if _, err := DB.Exec(createConcertTableSQL); err != nil {
		return fmt.Errorf("error running concert table migration: %w", err)
	}

	// Run Participant Migration
	if _, err := DB.Exec(createParticipantTableSQL); err != nil {
		return fmt.Errorf("error running concert table migration: %w", err)
	}

	// Run Payment Migration
	if _, err := DB.Exec(createPaymentTableSQL); err != nil {
		return fmt.Errorf("error running concert table migration: %w", err)
	}

	// Run Booking Migration
	if _, err := DB.Exec(createBookingTableSQL); err != nil {
		return fmt.Errorf("error running concert table migration: %w", err)
	}

	fmt.Println("Migrations completed successfully.")
	return nil
}
