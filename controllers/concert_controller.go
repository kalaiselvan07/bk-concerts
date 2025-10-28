package controllers

import (
	"io"
	"log"
	"net/http"
	"strings"

	"bk-concerts/applications/concert" // Using the correct module path

	"github.com/labstack/echo/v4"
)

// CreateConcertController handles POST requests to create a new concert.
func CreateConcertController(c echo.Context) error {
	// 1. Read the raw request body payload
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error reading payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload.",
		})
	}

	// 2. Call the Use Case function
	newConcert, err := concert.CreateConcert(payload)

	// 3. Handle errors (e.g., unmarshal failure, database insertion error)
	if err != nil {
		log.Printf("Concert creation failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create concert: " + err.Error(),
		})
	}

	// 4. Send a success response (201 Created) with the new object
	log.Println("Concert created successfully. ID:", newConcert.ConcertID)
	return c.JSON(http.StatusCreated, newConcert)
}

// GetConcertController handles GET requests to retrieve a single concert by ID.
func GetConcertController(c echo.Context) error {
	// 1. Extract the concertID from the path parameters
	concertID := c.Param("concertID")

	// 2. Call the use case function
	cData, err := concert.GetConcert(concertID)

	// 3. Handle errors
	if err != nil {
		log.Printf("Error fetching concert %s: %v", concertID, err)

		// A. Not Found Error (based on the use case error message)
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Concert not found.",
			})
		}

		// B. Invalid ID or other internal errors
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve concert: " + err.Error(),
		})
	}

	// 4. Return the retrieved concert object with a 200 OK status
	log.Println("Concert retrieved successfully. ID:", cData.ConcertID)
	return c.JSON(http.StatusOK, cData)
}

// GetAllConcertsController handles GET requests to retrieve all concerts.
func GetAllConcertsController(c echo.Context) error {
	// 1. Call the use case function
	concertsList, err := concert.GetAllConcerts()

	// 2. Handle errors
	if err != nil {
		log.Printf("Error fetching all concerts: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve concert data: " + err.Error(),
		})
	}

	// 3. Return the slice of concert objects with a 200 OK status.
	log.Println("Successfully retrieved all concerts.")
	return c.JSON(http.StatusOK, concertsList)
}

// DeleteConcertController handles DELETE requests to remove a single concert by ID.
func DeleteConcertController(c echo.Context) error {
	// 1. Extract the concertID from the path parameters
	concertID := c.Param("concertID")

	// 2. Call the use case function
	rowsAffected, err := concert.DeleteConcert(concertID)

	// 3. Handle errors
	if err != nil {
		log.Printf("Error deleting concert %s: %v", concertID, err)

		// If the concert was not found (based on the use case error message)
		if rowsAffected == 0 || strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Concert not found.",
			})
		}

		// Handle other internal errors
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete concert: " + err.Error(),
		})
	}

	// 4. Return 204 No Content for a successful deletion.
	log.Printf("Concert %s deleted successfully. Rows affected: %d", concertID, rowsAffected)

	// c.NoContent is the standard way to return 204.
	return c.NoContent(http.StatusNoContent)
}

// UpdateConcertController handles PUT requests to update concert details.
func UpdateConcertController(c echo.Context) error {
	seatID := c.Param("concertID")

	// 1. Read the raw request body payload
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error reading payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload."})
	}

	// 2. Call the Use Case
	cData, err := concert.UpdateConcert(seatID, payload)

	// 3. Handle errors
	if err != nil {
		log.Printf("Error updating concert %s: %v", seatID, err)

		// Check for specific use case errors:
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Concert not found."})
		}
		if strings.Contains(err.Error(), "invalid ID format") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// General Database/Unmarshalling Error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Concert update failed: " + err.Error()})
	}

	// 4. Send a success response (200 OK) with the updated details
	log.Println("Concert details updated successfully. ID:", cData.ConcertID)
	return c.JSON(http.StatusOK, cData)
}
