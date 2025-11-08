package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"supra/applications/auth"
	"supra/applications/user" // Assumes user use cases are available
	"supra/logger"            // ⬅️ Assuming this import path

	"github.com/labstack/echo/v4"
)

type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

// LoginHandler handles both Regular and Admin login flows.
func LoginHandler(c echo.Context) error {
	params := new(user.LoginParams)
	if err := c.Bind(params); err != nil {
		logger.Log.Warn(fmt.Sprintf("[auth] Login attempt failed: Invalid request binding for email: %s", params.Email))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login request"})
	}

	// Log the initiation of the login attempt
	logger.Log.Info(fmt.Sprintf("[auth] Login initiated for email: %s. Password provided: %t", params.Email, params.Password != ""))

	// 1. ADMIN LOGIN: If a password is provided, assume Admin login (high security flow)
	if params.Password != "" {
		logger.Log.Info(fmt.Sprintf("[auth] Attempting Admin login for email: %s", params.Email))

		token, role, err := auth.LoginAdmin(params.Email, params.Password)
		if err != nil {
			logger.Log.Warn(fmt.Sprintf("[auth] Admin login failed for %s: %v", params.Email, err))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials or user not found"})
		}

		logger.Log.Info(fmt.Sprintf("[auth] Admin login successful for %s. Role: %s", params.Email, role))
		return c.JSON(http.StatusOK, LoginResponse{Token: token, Role: role})
	}

	// 2. REGULAR USER LOGIN: Email-only -> Start OTP flow
	logger.Log.Info(fmt.Sprintf("[auth] Initiating OTP flow for regular user: %s", params.Email))

	token, role, err := auth.RequestUserOTP(params.Email)
	if err != nil {
		// Log the failure to initiate OTP (e.g., mail server failure, user creation failure)
		logger.Log.Error(fmt.Sprintf("[auth] Failed to initiate OTP for %s: %v", params.Email, err))

		if strings.Contains(err.Error(), "user not found") {
			// Although RequestUserOTP should handle creation, we log the severe error anyway.
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to initiate OTP process: " + err.Error()})
		}
		// If the error isn't critical (e.g., mail server is down), we may still proceed with a status message if appropriate,
		// but since the use case returns an error, we treat it as a failure to send.
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send OTP. Please try again."})
	}

	// If the user already has a token (i.e., this may signal a successful re-login without OTP re-send logic)
	if token != "" {
		logger.Log.Info(fmt.Sprintf("[auth] Regular user %s logged in successfully (re-login/token refresh). Role: %s", params.Email, role))
		return c.JSON(http.StatusOK, LoginResponse{Token: token, Role: role})
	}

	// Default success response for starting the OTP flow
	logger.Log.Info(fmt.Sprintf("[auth] OTP generated and request initiated for %s. Awaiting verification.", params.Email))
	return c.JSON(http.StatusOK, map[string]string{"message": "OTP sent. Check your email."})
}

// VerifyOTPHandler handles the second step of Regular User login.
func VerifyOTPHandler(c echo.Context) error {
	type VerifyParams struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	params := new(VerifyParams)
	if err := c.Bind(params); err != nil {
		logger.Log.Warn(fmt.Sprintf("[auth] OTP verification failed: Invalid request binding for email: %s", params.Email))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid verification request"})
	}

	logger.Log.Info(fmt.Sprintf("[auth] Attempting OTP verification for email: %s", params.Email))

	token, role, err := auth.VerifyOTP(params.Email, params.Code)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[auth] OTP verification failed for %s: %v", params.Email, err))
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired OTP."})
	}

	logger.Log.Info(fmt.Sprintf("[auth] OTP verified successfully for %s. JWT issued. Role: %s", params.Email, role))
	return c.JSON(http.StatusOK, LoginResponse{Token: token, Role: role})
}
