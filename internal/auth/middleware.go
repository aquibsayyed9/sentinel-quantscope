// internal/auth/middleware.go
package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// func AuthMiddleware(tokenService TokenService) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
// 			c.Abort()
// 			return
// 		}

// 		// Check if the Authorization header has the right format
// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header format must be Bearer {token}"})
// 			c.Abort()
// 			return
// 		}

// 		// Validate the token
// 		claims, err := tokenService.ValidateToken(parts[1])
// 		if err != nil {
// 			var status int
// 			var message string

// 			switch err {
// 			case ErrExpiredToken:
// 				status = http.StatusUnauthorized
// 				message = "token has expired"
// 			case ErrInvalidToken:
// 				status = http.StatusUnauthorized
// 				message = "invalid token"
// 			default:
// 				status = http.StatusInternalServerError
// 				message = "failed to validate token"
// 			}

// 			c.JSON(status, gin.H{"error": message})
// 			c.Abort()
// 			return
// 		}

// 		// Set user information in the context
// 		c.Set("userID", claims.UserID)
// 		c.Set("email", claims.Email)

// 		c.Next()
// 	}
// }

func AuthMiddleware(tokenService TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Check if the Authorization header has the right format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := tokenService.ValidateToken(parts[1])
		if err != nil {
			// For any token validation error, return 401 Unauthorized
			// Only use 500 for unexpected server errors, not validation failures
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}
