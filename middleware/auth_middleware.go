package middleware

import (
	"net/http"
	"strings"

	"github.com/ShreyasDr71/GoPAMS/database"
	"github.com/ShreyasDr71/GoPAMS/models"
	"github.com/ShreyasDr71/GoPAMS/services"
	"github.com/gin-gonic/gin"
)

// AuthRequired checks if the request is authenticated with a valid JWT token.
// It also enforces the "forced password change on first login" policy.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		// 1. Try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenStr = parts[1]
			}
		}

		// 2. Try to get token from cookie if header is empty
		if tokenStr == "" {
			if cookie, err := c.Cookie("token"); err == nil {
				tokenStr = cookie
			}
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authentication token required"})
			c.Abort()
			return
		}

		// 3. Validate token
		claims, err := services.ValidateJWT(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 4. Fetch latest user state from database
		var user models.User
		if err := database.DB.Preload("Role").Preload("Group").First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not found"})
			c.Abort()
			return
		}

		// 5. Inject user and claims into context
		c.Set("currentUser", &user)
		c.Set("claims", claims)

		// 6. Enforce forced password change on first login
		// Users with MustChangePassword = true can ONLY access /api/auth/change-password and /api/auth/logout.
		if user.MustChangePassword {
			path := c.Request.URL.Path
			if path != "/api/auth/change-password" && path != "/api/auth/logout" && path != "/api/auth/me" {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "password_change_required",
					"message": "First login password change is required",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
