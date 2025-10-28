package main

import (
	"bk-concerts/controllers"
	"bk-concerts/db"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// main method
	e := echo.New()
	e.Use(middleware.Logger())

	connStr := "user=postgres password=postgres dbname=postgres sslmode=disable"

	if err := db.InitDB(connStr); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.DB.Close()

	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// Seat routes
	e.POST("/seats", controllers.AddSeatHandler)
	e.GET("/seats/:seatID", controllers.GetSeatHandler)
	e.DELETE("/seats/:seatID", controllers.DeleteSeatHandler)
	e.GET("/seats", controllers.GetAllSeatsHandler)
	e.PUT("/seats/:seatID", controllers.UpdateSeatController)

	// Concert routes
	e.POST("/concerts", controllers.CreateConcertController)
	e.GET("/concerts/:concertID", controllers.GetConcertController)
	e.GET("/concerts", controllers.GetAllConcertsController)
	e.DELETE("/concerts/:concertID", controllers.DeleteConcertController)
	e.PUT("/concerts/:concertID", controllers.UpdateConcertController)

	// Participant routes
	e.POST("/participants", controllers.AddParticipantController)
	e.GET("/participants", controllers.GetAllParticipantsController)
	e.GET("/participants/:userID", controllers.GetParticipantController)
	e.PUT("/participants/:userID", controllers.UpdateParticipantController)
	e.DELETE("/participants/:userID", controllers.DeleteParticipantController)

	// Payment routes
	e.POST("/payments", controllers.AddPaymentController)
	e.GET("/payments", controllers.GetAllPaymentsController)
	e.GET("/payments/:paymentID", controllers.GetPaymentController)
	e.PUT("/payments/:paymentID", controllers.UpdatePaymentController)
	e.DELETE("/payments/:paymentID", controllers.DeletePaymentController)

	// Booking routes
	e.POST("/bookings", controllers.BookNowController)
	e.GET("/bookings", controllers.GetAllBookingsController)
	e.GET("/bookings/:bookingID", controllers.GetBookingController)
	e.PUT("/bookings/:bookingID", controllers.UpdateBookingController)
	e.DELETE("/bookings/:bookingID", controllers.DeleteBookingController)

	// 4. Start the server
	log.Println("Starting Echo server on http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
