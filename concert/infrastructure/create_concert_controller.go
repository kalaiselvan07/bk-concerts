package infrastructure

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"supra/concert/application"

	"github.com/labstack/echo/v4"
)

type CreateConcertController struct {
	log *slog.Logger
	uc  *application.CreateConcertUC
}

func NewCreateConcertController(log *slog.Logger) *CreateConcertController {
	return &CreateConcertController{
		log: log,
		uc:  application.NewCreateConcertUC(log),
	}
}

// CreateConcertController handles POST requests to create a new concert.
func (c *CreateConcertController) Invoke(ctx echo.Context) error {
	// 1. Read the raw request body payload
	payload, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		log.Printf("Error reading payload: %v", err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload.",
		})
	}

	// 2. Call the Use Case function
	newConcert, err := c.uc.Invoke(payload)

	// 3. Handle errors (e.g., unmarshal failure, database insertion error)
	if err != nil {
		log.Printf("Concert creation failed: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create concert: " + err.Error(),
		})
	}

	// 4. Send a success response (201 Created) with the new object
	log.Println("Concert created successfully. ID:", newConcert.ConcertID)
	return ctx.JSON(http.StatusCreated, newConcert)
}
