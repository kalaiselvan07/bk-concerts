package infrastructure

import (
	"log"
	"log/slog"
	"net/http"
	"strings"

	"supra/concert/application"

	"github.com/labstack/echo/v4"
)

type DeleteConcertController struct {
	log *slog.Logger
	uc  *application.DeleteConcertUC
}

func NewDeleteConcertController(log *slog.Logger) *DeleteConcertController {
	return &DeleteConcertController{
		log: log,
		uc:  application.NewDeleteConcertUC(log),
	}
}

// DeleteConcertController handles DELETE requests to remove a single concert by ID.
func (c *DeleteConcertController) Invoke(ctx echo.Context) error {
	// 1. Extract the concertID from the path parameters
	concertID := ctx.Param("concertID")

	// 2. Call the use case function
	rowsAffected, err := c.uc.Invoke(concertID)

	// 3. Handle errors
	if err != nil {
		log.Printf("Error deleting concert %s: %v", concertID, err)

		// If the concert was not found (based on the use case error message)
		if rowsAffected == 0 || strings.Contains(err.Error(), "not found") {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Concert not found.",
			})
		}

		// Handle other internal errors
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete concert: " + err.Error(),
		})
	}

	// 4. Return 204 No Content for a successful deletion.
	log.Printf("Concert %s deleted successfully. Rows affected: %d", concertID, rowsAffected)

	// c.NoContent is the standard way to return 204.
	return ctx.NoContent(http.StatusNoContent)
}
