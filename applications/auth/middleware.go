package auth

import (
	"fmt"
	"net/http"
	"strings"

	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// NOTE: UserClaims struct and jwtSecret variable are assumed to be
// defined in another file within this 'auth' package (e.g., auth.go).

func JWTAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		tokenString := ""

		// 1️⃣ Prefer Authorization header
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 2️⃣ Fallback: Check ?token= in query (useful for email verification links)
		if tokenString == "" {
			tokenString = c.QueryParam("token")
			if tokenString != "" {
				logger.Log.Info(fmt.Sprintf("[auth] Using token from query parameter for path: %s", c.Path()))
			}
		}

		if tokenString == "" {
			logger.Log.Warn("[auth] JWT check failed: No token in header or query.")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization token missing"})
		}

		// 3️⃣ Validate the token
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			logger.Log.Warn(fmt.Sprintf("[auth] Invalid or expired JWT: %v", err))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			logger.Log.Error("[auth] JWT claims extraction failed")
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid token claims"})
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("userEmail", claims.Email)

		logger.Log.Info(fmt.Sprintf("[auth] ✅ JWT validated. UserID: %s, Role: %s", claims.UserID, claims.Role))
		return next(c)
	}
}

// AdminOnlyMiddleware checks the role set by JWTAuthMiddleware.
func AdminOnlyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get("userRole")

		// Log the role being checked
		logger.Log.Info(fmt.Sprintf("[auth] RBAC check for path %s. Required role: Admin. User role: %v", c.Path(), role))

		// Ensure the token validation ran and set the role
		if role == nil || role != "admin" {
			logger.Log.Warn(fmt.Sprintf("[auth] RBAC FAILED for UserID %v: Access Forbidden.", c.Get("userID")))
			// Block access if role is missing or not "admin"
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Forbidden: Admin privileges required"})
		}

		logger.Log.Info(fmt.Sprintf("[auth] RBAC PASSED for Admin UserID: %v", c.Get("userID")))
		return next(c)
	}
}
