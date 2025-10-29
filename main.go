package main

import (
	"bk-concerts/applications/auth"
	"bk-concerts/controllers"
	"bk-concerts/db"
	"bk-concerts/logger" // Your logging package
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file. Continuing...")
	}

	e := echo.New()

	// --- INITIAL STARTUP LOGGING ---
	logger.Log.Info("[main] program started")
	logger.Log.Info("[main] Configuring global middleware and database connection.")

	// Global Middleware: Logger and CORS (CRITICAL for frontend connection)
	e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://bk-concerts-ui.vercel.app",
		},
		AllowMethods: []string{
			echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization,
		},
	}))

	// --- DATABASE CONNECTION ---
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// local fallback (what you had)
		dsn = "user=postgres password=postgres dbname=postgres sslmode=disable"
		logger.Log.Warn("[main] DATABASE_URL not set. Falling back to local Postgres.")
	} else {
		// Ensure sslmode=require for Neon (safeguard if the copied URL lacked it)
		if !strings.Contains(dsn, "sslmode=") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn = dsn + sep + "sslmode=require"
		}
	}

	logger.Log.Info("[main] Attempting to connect to PostgreSQL...")
	if err := db.InitDB(dsn); err != nil {
		logger.Log.Error(fmt.Sprintf("[main] Database connection failed: %v", err))
		log.Fatalf("Database initialization failed: %v", err)
	}
	logger.Log.Info("[main] Database connection successful.")
	defer func() {
		_ = db.DB.Close()
		logger.Log.Info("[main] Database connection closed.")
	}()

	// Optional: tune connection pool for serverless Postgres like Neon
	db.DB.SetMaxOpenConns(5)
	db.DB.SetMaxIdleConns(5)
	db.DB.SetConnMaxLifetime(30 * time.Minute)

	// --- MIGRATIONS ---
	logger.Log.Info("[main] Running database migrations...")
	if err := db.RunMigrations(); err != nil {
		logger.Log.Error(fmt.Sprintf("[main] Database migration failed: %v", err))
		log.Fatalf("Database migration failed: %v", err)
	}
	logger.Log.Info("[main] Database migrations completed successfully.")

	// --- 1. PUBLIC ROUTES (No Auth Required) ---
	logger.Log.Info("[router] Registering public authentication and read-only routes.")

	// Authentication/Login routes
	e.POST("/login", controllers.LoginHandler)
	e.POST("/verify-otp", controllers.VerifyOTPHandler)

	// --- 2. PROTECTED GROUP (Requires Valid JWT Token) ---
	logger.Log.Info("[router] Configuring '/api/v1' protected group (JWT Required).")

	r := e.Group("/api/v1")
	r.Use(auth.JWTAuthMiddleware)

	// Booking Routes (Making a booking, viewing history)
	r.POST("/bookings", controllers.BookNowController)
	// we'll create a new api to list user specific history not all booking

	// --- 3. ADMIN-ONLY GROUP (Requires JWT + Admin Role) ---
	logger.Log.Warn("[router] Configuring '/api/v1/admin' group (Admin Role Required).")

	admin := r.Group("/admin")
	admin.Use(auth.AdminOnlyMiddleware)

	// --- ADMIN CRUD ROUTES ---

	// Seats
	r.GET("/seats", controllers.GetAllSeatsHandler)
	r.GET("/seats/:seatID", controllers.GetSeatHandler)
	admin.POST("/seats", controllers.AddSeatHandler)
	admin.PUT("/seats/:seatID", controllers.UpdateSeatController)
	admin.DELETE("/seats/:seatID", controllers.DeleteSeatHandler)
	logger.Log.Info("[router] Admin: Seats CRUD configured.")

	// Concerts
	r.GET("/concerts", controllers.GetAllConcertsController) // NOTE: Corrected GetAllSeatsHandler to GetAllConcertsController
	r.GET("/concerts/:concertID", controllers.GetConcertController)
	admin.POST("/concerts", controllers.CreateConcertController)
	admin.PUT("/concerts/:concertID", controllers.UpdateConcertController)
	admin.DELETE("/concerts/:concertID", controllers.DeleteConcertController)
	logger.Log.Info("[router] Admin: Concerts CRUD configured.")

	// Participants
	r.GET("/participants", controllers.GetAllParticipantsController)
	r.GET("/participants/:userID", controllers.GetParticipantController)
	admin.POST("/participants", controllers.AddParticipantController)
	admin.PUT("/participants/:userID", controllers.UpdateParticipantController)
	admin.DELETE("/participants/:userID", controllers.DeleteParticipantController)
	logger.Log.Info("[router] Admin: Participants CRUD configured.")

	// Payments
	r.GET("/payments", controllers.GetAllPaymentsController)
	r.GET("/payments/:paymentID", controllers.GetPaymentController)
	admin.POST("/payments", controllers.AddPaymentController)
	admin.PUT("/payments/:paymentID", controllers.UpdatePaymentController)
	admin.DELETE("/payments/:paymentID", controllers.DeletePaymentController)
	logger.Log.Info("[router] Admin: Payments CRUD configured.")

	// Booking Update/Delete
	r.GET("/bookings", controllers.GetAllBookingsController)
	r.GET("/bookings/:bookingID", controllers.GetBookingController)
	r.PATCH("/bookings/:bookingID/receipt", controllers.UpdateBookingReceiptController)
	r.GET("/bookings/:bookingID/receipt", controllers.GetBookingReceiptController)
	admin.PUT("/bookings/:bookingID", controllers.UpdateBookingController)
	admin.DELETE("/bookings/:bookingID", controllers.DeleteBookingController)
	admin.GET("/bookings", controllers.GetAllBookingsAdminController)
	admin.GET("/bookings/pending", controllers.GetPendingBookingsController)
	admin.PATCH("/bookings/:bookingID/verify", controllers.VerifyBookingController)

	logger.Log.Info("[router] Admin: Booking Update/Delete configured.")

	// 4. Start the server
	log.Println("Starting Echo server on http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
