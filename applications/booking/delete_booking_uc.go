package booking

import (
	"context"
	"database/sql"
	"fmt"

	"bk-concerts/applications/seat"
	"bk-concerts/db"
)

// DeleteBooking handles the cancellation logic: changing status and refunding seats.
// We change the status to CANCELLED instead of deleting the row for audit purposes.
func DeleteBooking(bookingID string) (*Booking, error) {
	// Start a transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Fetch the existing booking details (and lock the row for update)
	currentBooking, err := GetBookingTx(tx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking retrieval failed: %w", err)
	}

	// Prevent cancellation if already cancelled or confirmed (business logic decision)
	if currentBooking.BookingStatus == CANCELLED {
		tx.Rollback()
		return nil, fmt.Errorf("booking ID %s is already cancelled", bookingID)
	}

	// 2. "Refund" the seats (Increase the available count)

	// Check the current seat availability and update within the transaction.
	// We need a helper function in the seat package that handles the addition.
	// For simplicity, we assume seat.IncreaseAvailableTx exists (similar to UpdateAvailableTx).

	_, err = increaseSeatAvailabilityTx(tx, currentBooking.SeatID, currentBooking.SeatQuantity)
	if err != nil {
		return nil, fmt.Errorf("failed to refund seats: %w", err)
	}

	// 3. Update Booking Status to CANCELLED
	const updateSQL = `
		UPDATE booking
		SET booking_status = $2
		WHERE booking_id = $1
		RETURNING booking_id, booking_email, booking_status, payment_details_id, 
				  receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
				  participant_ids, created_at`

	// Note: We reuse the RETURNING statement logic from UpdateBooking for convenience
	updatedBk := &Booking{}
	var receiptImage []byte
	var participantIDsJSON []byte

	row := tx.QueryRow(updateSQL, currentBooking.BookingID, CANCELLED)

	if err := row.Scan(
		&updatedBk.BookingID, &updatedBk.BookingEmail, &updatedBk.BookingStatus, &updatedBk.PaymentDetailsID,
		&receiptImage, &updatedBk.SeatQuantity, &updatedBk.SeatID, &updatedBk.TotalAmount, &updatedBk.SeatType,
		&participantIDsJSON, &updatedBk.CreatedAt,
	); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to scan cancelled booking: %w", err)
	}

	// 4. Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit cancellation: %w", err)
	}

	return updatedBk, nil
}

// ⚠️ REQUIRED TRANSACTIONAL HELPER IN SEAT PACKAGE ⚠️
// This function needs to be added to the seat package to compile.
func increaseSeatAvailabilityTx(tx *sql.Tx, seatID string, quantity int) (*seat.Seat, error) {
	// Logic:
	// 1. Get current seat FOR UPDATE (to lock the row).
	currentSeat, err := seat.GetSeatForUpdateTx(tx, seatID)
	if err != nil {
		return nil, err
	}
	// 2. Calculate new total: current + refunded quantity
	newVal := currentSeat.Available + quantity

	// 3. Update the available count in the database.
	updatedSeat, err := seat.UpdateAvailableTx(tx, seatID, newVal)
	if err != nil {
		return nil, fmt.Errorf("failed to increase seat count: %w", err)
	}
	return updatedSeat, nil
}
