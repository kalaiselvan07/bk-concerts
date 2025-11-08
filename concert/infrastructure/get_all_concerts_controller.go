package infrastructure

import (
	"log"
	"log/slog"
	"net/http"
	"supra/concert/application"

	"github.com/labstack/echo/v4"
)

type GetAllConcertsController struct {
	log *slog.Logger
	uc  *application.GetAllConcertsUC
}

func NewGetAllConcertsController(log *slog.Logger) *GetAllConcertsController {
	return &GetAllConcertsController{
		log: log,
		uc:  application.NewGetAllConcertsUC(log),
	}
}

func (c *GetAllConcertsController) Invoke(ctx echo.Context) error {
	// 1. Call the use case function
	concertsList, err := c.uc.Invoke()

	// 2. Handle errors
	if err != nil {
		log.Printf("Error fetching all concerts: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve concert data: " + err.Error(),
		})
	}

	// 3. Return the slice of concert objects with a 200 OK status.
	log.Println("Successfully retrieved all concerts.")
	return ctx.JSON(http.StatusOK, concertsList)
}
