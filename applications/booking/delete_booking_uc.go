package booking

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"supra/applications/seat"
	"supra/db"
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
	// NOTE: Booking struct, constants, and GetBookingTx assumed here
)

// DeleteBooking handles the cancellation logic: changing status and refunding seats.
// We change the status to CANCELLED instead of deleting the row for audit purposes.
func DeleteBooking(bookingID string) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[delete-booking-uc] Starting cancellation process for BookingID: %s", bookingID))

	// Start a transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-booking-uc] Failed to start transaction for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	logger.Log.Info(fmt.Sprintf("[delete-booking-uc] Transaction started for %s.", bookingID))

	// 1. Fetch the existing booking details (and lock the row for update)
	currentBooking, err := GetBookingTx(tx, bookingID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-booking-uc] Booking retrieval failed for %s: %v", bookingID, err))
		return nil, fmt.Errorf("booking retrieval failed: %w", err)
	}

	// Prevent cancellation if already cancelled or confirmed (business logic decision)
	if currentBooking.BookingStatus == CANCELLED {
		tx.Rollback()
		logger.Log.Warn(fmt.Sprintf("[delete-booking-uc] Cancellation skipped for %s: Already CANCELLED.", bookingID))
		return nil, fmt.Errorf("booking ID %s is already cancelled", bookingID)
	}
	logger.Log.Info(fmt.Sprintf("[delete-booking-uc] Booking %s found (Status: %s). Proceeding to refund seats.", bookingID, currentBooking.BookingStatus))

	// 2. "Refund" the seats (Increase the available count)
	// We call the helper which uses the seat package functions within the transaction.
	_, err = increaseSeatAvailabilityTx(tx, currentBooking.SeatID, currentBooking.SeatQuantity)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-booking-uc] Seat refund failed for %s (Rollback): %v", bookingID, err))
		return nil, fmt.Errorf("failed to refund seats: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[delete-booking-uc] Successfully refunded %d seats to SeatID %s.", currentBooking.SeatQuantity, currentBooking.SeatID))

	// 3. Update Booking Status to CANCELLED
	const updateSQL = `
		UPDATE booking
		SET booking_status = $2
		WHERE booking_id = $1
		RETURNING booking_id, booking_email, booking_status, payment_details_id, 
				  receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
				  participant_ids, created_at, user_notes`

	// Note: We reuse the RETURNING statement logic from UpdateBooking for convenience
	updatedBk := &Booking{}
	var receiptImage []byte
	var participantIDsJSON []byte
	var bookingIDUUID uuid.UUID // Helper for scanning UUID

	row := tx.QueryRow(updateSQL, currentBooking.BookingID, CANCELLED)

	if err := row.Scan(
		&bookingIDUUID, &updatedBk.BookingEmail, &updatedBk.BookingStatus, &updatedBk.PaymentDetailsID,
		&receiptImage, &updatedBk.SeatQuantity, &updatedBk.SeatID, &updatedBk.TotalAmount, &updatedBk.SeatType,
		&participantIDsJSON, &updatedBk.CreatedAt, &updatedBk.UserNotes,
	); err != nil {
		tx.Rollback()
		logger.Log.Error(fmt.Sprintf("[delete-booking-uc] Failed to scan RETURNING row after status update (Rollback): %v", err))
		return nil, fmt.Errorf("failed to scan cancelled booking: %w", err)
	}
	updatedBk.BookingID = bookingIDUUID // Map UUID

	// 4. Commit the transaction
	if err := tx.Commit(); err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-booking-uc] Failed to commit transaction for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to commit cancellation: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[delete-booking-uc] Booking %s successfully committed as CANCELLED.", bookingID))

	// Finalize mapping for the return struct
	updatedBk.ReceiptImage = receiptImage
	if len(participantIDsJSON) > 0 && string(participantIDsJSON) != "null" {
		json.Unmarshal(participantIDsJSON, &updatedBk.ParticipantIDs)
	}

	return updatedBk, nil
}

// ⚠️ REQUIRED TRANSACTIONAL HELPER IN SEAT PACKAGE ⚠️
// This function needs to be added to the seat package to compile.
func increaseSeatAvailabilityTx(tx *sql.Tx, seatID string, quantity int) (*seat.Seat, error) {
	logger.Log.Info(fmt.Sprintf("[delete-booking-uc] Initiating seat count increase for SeatID %s by %d.", seatID, quantity))

	// Logic:
	// 1. Get current seat FOR UPDATE (to lock the row).
	currentSeat, err := seat.GetSeatForUpdateTx(tx, seatID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-booking-uc] Failed to lock seat %s for refund: %v", seatID, err))
		return nil, err
	}
	// 2. Calculate new total: current + refunded quantity
	newVal := currentSeat.Available + quantity

	// 3. Update the available count in the database.
	updatedSeat, err := seat.UpdateAvailableTx(tx, seatID, newVal)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-booking-uc] Failed to update seat count for refund %s: %v", seatID, err))
		return nil, fmt.Errorf("failed to increase seat count: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[delete-booking-uc] Seat %s count increased successfully to %d.", seatID, newVal))

	return updatedSeat, nil
}
