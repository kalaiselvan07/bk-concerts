package controllers

import (
	"io"
	"log"
	"net/http"
	"strings"

	"bk-concerts/applications/participant" // Using the correct module path

	"github.com/labstack/echo/v4"
)

// AddParticipantController handles POST requests to create a new participant.
func AddParticipantController(c echo.Context) error {
	// 1. Read the raw request body payload
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error reading payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload.",
		})
	}

	// 2. Call the Use Case function
	newParticipant, err := participant.AddParticipant(payload)

	// 3. Handle errors
	if err != nil {
		log.Printf("Participant creation failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create participant: " + err.Error(),
		})
	}

	// 4. Send a success response (201 Created)
	log.Println("Participant created successfully. ID:", newParticipant.UserID)
	return c.JSON(http.StatusCreated, newParticipant)
}

// GetParticipantController handles GET /participants/:userID (Read One)
func GetParticipantController(c echo.Context) error {
	userID := c.Param("userID")
	p, err := participant.GetParticipant(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Participant not found."})
		}
		if strings.Contains(err.Error(), "invalid user ID format") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve participant: " + err.Error()})
	}
	return c.JSON(http.StatusOK, p)
}

// GetAllParticipantsController handles GET /participants (Read All)
func GetAllParticipantsController(c echo.Context) error {
	participantsList, err := participant.GetAllParticipants()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve participants data: " + err.Error()})
	}
	return c.JSON(http.StatusOK, participantsList)
}

// UpdateParticipantController handles PUT /participants/:userID (Update Details)
func UpdateParticipantController(c echo.Context) error {
	userID := c.Param("userID")
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload."})
	}

	p, err := participant.UpdateParticipant(userID, payload)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Participant not found."})
		}
		if strings.Contains(err.Error(), "invalid user ID format") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Participant update failed: " + err.Error()})
	}
	return c.JSON(http.StatusOK, p)
}

// DeleteParticipantController handles DELETE /participants/:userID (Delete)
func DeleteParticipantController(c echo.Context) error {
	userID := c.Param("userID")
	rowsAffected, err := participant.DeleteParticipant(userID)
	if err != nil {
		if rowsAffected == 0 || strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Participant not found."})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete participant: " + err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
