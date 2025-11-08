package infrastructure

import (
	"log"
	"log/slog"
	"net/http"
	"strings"

	"supra/concert/application"

	"github.com/labstack/echo/v4"
)

type GetConcertByIDController struct {
	l  *slog.Logger
	uc *application.GetConcertByIDUC
}

func NewGetConcertByIDController(l *slog.Logger) *GetConcertByIDController {
	return &GetConcertByIDController{
		l:  l,
		uc: application.NewGetConcertByIDUC(l),
	}
}

func (c *GetConcertByIDController) Invoke(ctx echo.Context) error {
	// 1. Extract the concertID from the path parameters
	concertID := ctx.Param("concertID")

	// 2. Call the use case function
	cData, err := c.uc.Invoke(concertID)

	// 3. Handle errors
	if err != nil {
		log.Printf("Error fetching concert %s: %v", concertID, err)

		// A. Not Found Error (based on the use case error message)
		if strings.Contains(err.Error(), "not found") {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Concert not found.",
			})
		}

		// B. Invalid ID or other internal errors
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve concert: " + err.Error(),
		})
	}

	// 4. Return the retrieved concert object with a 200 OK status
	log.Println("Concert retrieved successfully. ID:", cData.ConcertID)
	return ctx.JSON(http.StatusOK, cData)
}
