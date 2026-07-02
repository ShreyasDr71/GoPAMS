package handlers

import (
	"net/http"

	"github.com/ShreyasDr71/GoPAMS/config"
	"github.com/ShreyasDr71/GoPAMS/database"
	"github.com/ShreyasDr71/GoPAMS/models"
	"github.com/ShreyasDr71/GoPAMS/services"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// Login handles POST /api/auth/login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Username and password are required"})
		return
	}

	var user models.User
	if err := database.DB.Preload("Role").Preload("Group").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid username or password"})
		return
	}

	if !services.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid username or password"})
		return
	}

	roleName := "Guest"
	if user.Role != nil {
		roleName = user.Role.Name
	}

	token, err := services.GenerateJWT(user.ID, user.Username, user.IsAdmin, user.MustChangePassword, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to generate session"})
		return
	}

	// Set HttpOnly cookie for safety
	cookieMaxAge := config.AppConfig.SessionTimeoutMinutes * 60
	c.SetCookie("token", token, cookieMaxAge, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message":              "Login successful",
		"token":                token,
		"must_change_password": user.MustChangePassword,
		"user": gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"full_name":   user.FullName,
			"is_admin":    user.IsAdmin,
			"phone":       user.PhoneNumber,
			"email":       user.Email,
			"employee_id": user.EmployeeID,
			"role":        roleName,
			"group":       user.Group,
		},
	})
}

// ChangePassword handles POST /api/auth/change-password
func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Current and new passwords are required"})
		return
	}

	// Password complexity check
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "New password must be at least 8 characters long"})
		return
	}

	val, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Not logged in"})
		return
	}
	user := val.(*models.User)

	// Verify old password
	if !services.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Incorrect current password"})
		return
	}

	// Ensure new password is not the same as the old one
	if req.CurrentPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "New password must be different from current password"})
		return
	}

	hashedPassword, err := services.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to encrypt new password"})
		return
	}

	// Update user password
	user.PasswordHash = hashedPassword
	user.MustChangePassword = false

	if err := database.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to update password"})
		return
	}

	roleName := "Guest"
	if user.Role != nil {
		roleName = user.Role.Name
	}

	// Generate new token reflecting mustChangePassword = false
	newToken, err := services.GenerateJWT(user.ID, user.Username, user.IsAdmin, false, roleName)
	if err == nil {
		cookieMaxAge := config.AppConfig.SessionTimeoutMinutes * 60
		c.SetCookie("token", newToken, cookieMaxAge, "/", "", false, true)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
		"token":   newToken,
	})
}

// Logout handles POST /api/auth/logout
func Logout(c *gin.Context) {
	// Expire the cookie immediately
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetMe handles GET /api/auth/me
func GetMe(c *gin.Context) {
	val, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Not logged in"})
		return
	}
	user := val.(*models.User)

	roleName := "Guest"
	if user.Role != nil {
		roleName = user.Role.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                   user.ID,
			"username":             user.Username,
			"full_name":            user.FullName,
			"is_admin":             user.IsAdmin,
			"phone":                user.PhoneNumber,
			"email":                user.Email,
			"employee_id":          user.EmployeeID,
			"role":                 roleName,
			"group":                user.Group,
			"must_change_password": user.MustChangePassword,
			"created_at":           user.CreatedAt,
		},
	})
}
