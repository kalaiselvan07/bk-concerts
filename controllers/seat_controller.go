package controllers

import (
	"io"
	"log"
	"net/http"
	"strings"

	"supra/applications/seat"

	"github.com/labstack/echo/v4"
)

func AddSeatHandler(c echo.Context) error {
	// 1. Read all bytes from the request body
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Println("Failed to read request body:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	st, err := seat.AddSeat(payload)
	if err != nil {
		log.Println("Seat creation failed:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create seat: " + err.Error(),
		})
	}

	log.Println("Seat created successfully. ID:", st.SeatID)

	// ⬅️ THIS IS THE CORRECT SYNTAX FOR ECHO ⬅️
	return c.JSON(http.StatusCreated, st)
}

func GetSeatHandler(c echo.Context) error {
	// 1. Extract the seatID from the path parameters
	seatID := c.Param("seatID")

	// Basic check for empty ID
	if seatID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Seat ID is required.",
		})
	}

	// 2. Call the use case function
	st, err := seat.GetSeat(seatID)
	// 3. Handle errors from the use case (e.g., UUID parsing or DB errors)
	if err != nil {
		log.Printf("Error fetching seat %s: %v", seatID, err)

		// Check for the "Not Found" error specifically (as implemented in your GetSeat logic)
		// If the error message contains the "not found" phrase, return 404.
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Seat not found.",
			})
		}

		// Handle general internal errors (e.g., database connection failure)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve seat: " + err.Error(),
		})
	}

	// 4. Return the retrieved seat object (st) with a 200 OK status
	log.Println("Seat retrieved successfully. ID:", st.SeatID)
	return c.JSON(http.StatusOK, st)
}

func DeleteSeatHandler(c echo.Context) error {
	// 1. Extract the seatID from the path parameters
	seatID := c.Param("seatID")

	// 2. Call the use case function
	rowsAffected, err := seat.DeleteSeat(seatID)

	// 3. Handle errors
	if err != nil {
		log.Printf("Error deleting seat %s: %v", seatID, err)

		// If the seat was not found (based on the use case error message)
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Seat not found.",
			})
		}

		// Handle other internal errors (e.g., database connection failure)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete seat: " + err.Error(),
		})
	}

	// 4. Return 204 No Content for a successful deletion.
	log.Printf("Seat %s deleted successfully. Rows affected: %d", seatID, rowsAffected)

	// c.NoContent is used to send a response with no body.
	return c.NoContent(http.StatusNoContent)
}

// GetAllSeatsHandler handles GET requests to retrieve all seats.
func GetAllSeatsHandler(c echo.Context) error {
	// 1. Call the use case function
	seats, err := seat.GetAllSeats()

	// 2. Handle errors
	if err != nil {
		log.Printf("Error fetching all seats: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve seats data: " + err.Error(),
		})
	}

	// 3. Return the slice of seat objects with a 200 OK status.
	// Note: If no seats are found, it will return an empty JSON array ([]),
	// which is the correct HTTP response for an empty collection.
	log.Println("Successfully retrieved all seats.")
	return c.JSON(http.StatusOK, seats)
}

// UpdateSeatController handles PUT/PATCH requests to update seat details.
func UpdateSeatController(c echo.Context) error {
	seatID := c.Param("seatID")

	// 1. Read the raw request body payload
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error reading payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload."})
	}

	// 2. Call the Use Case with the path ID and payload
	// st is the updated *seat.Seat object returned from the DB.
	st, err := seat.UpdateSeat(seatID, payload)

	// 3. Handle errors
	if err != nil {
		log.Printf("Error updating seat %s: %v", seatID, err)

		// Check for specific use case errors:

		// A. Not Found Error (from the database check inside UpdateSeat)
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Seat not found."})
		}

		// B. Invalid ID Error (from UUID parsing)
		if strings.Contains(err.Error(), "invalid seat ID format") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// C. General Database/Unmarshalling Error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Seat update failed: " + err.Error()})
	}

	// 4. Send a success response (200 OK) with the updated details
	log.Println("Seat details updated successfully. ID:", st.SeatID)
	return c.JSON(http.StatusOK, st)
}
