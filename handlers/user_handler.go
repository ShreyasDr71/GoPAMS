package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/ShreyasDr71/GoPAMS/config"
	"github.com/ShreyasDr71/GoPAMS/database"
	"github.com/ShreyasDr71/GoPAMS/models"
	"github.com/ShreyasDr71/GoPAMS/services"
	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	FullName     string  `json:"full_name" binding:"required"`
	Username     string  `json:"username" binding:"required"`
	PhoneNumber  string  `json:"phone_number"`
	Email        *string `json:"email"`
	EmployeeID   *string `json:"employee_id"`
	GroupID      *uint   `json:"group_id"`
	RoleID       *uint   `json:"role_id"`
	IsAdmin      bool    `json:"is_admin"`
	TempPassword string  `json:"temp_password"`
}

type UpdateUserRequest struct {
	FullName    string  `json:"full_name" binding:"required"`
	PhoneNumber string  `json:"phone_number"`
	Email       *string `json:"email"`
	EmployeeID  *string `json:"employee_id"`
	GroupID     *uint   `json:"group_id"`
	RoleID      *uint   `json:"role_id"`
	IsAdmin     bool    `json:"is_admin"`
}

type ResetPasswordRequest struct {
	TempPassword string `json:"temp_password"`
}

func generateTempPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = charset[0]
		} else {
			b[i] = charset[n.Int64()]
		}
	}
	return string(b)
}

// ListUsers handles GET /api/users
func ListUsers(c *gin.Context) {
	var users []models.User
	if err := database.DB.Preload("Role").Preload("Group").Order("id asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to retrieve users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUser handles GET /api/users/:id
func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.Preload("Role").Preload("Group").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /api/users
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request fields"})
		return
	}

	// Enterprise Mode requirement check
	if config.AppConfig.EnterpriseMode {
		if req.EmployeeID == nil || *req.EmployeeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Employee ID is required in Enterprise Mode"})
			return
		}
	}

	// Validate username unique check
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Username already exists"})
		return
	}

	// Validate Group exists if provided
	if req.GroupID != nil {
		var g int64
		if err := database.DB.Model(&models.Group{}).Where("id = ?", req.GroupID).Count(&g).Error; err != nil || g == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Specified Group does not exist"})
			return
		}
	}

	// Validate Role exists if provided
	if req.RoleID != nil {
		var r int64
		if err := database.DB.Model(&models.Role{}).Where("id = ?", req.RoleID).Count(&r).Error; err != nil || r == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Specified Role does not exist"})
			return
		}
	}

	// Determine temporary password
	plaintextPass := req.TempPassword
	if plaintextPass == "" {
		plaintextPass = generateTempPassword(12)
	}

	// Hash password
	hashedPass, err := services.HashPassword(plaintextPass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to encrypt temporary password"})
		return
	}

	user := models.User{
		FullName:           req.FullName,
		Username:           req.Username,
		PasswordHash:       hashedPass,
		PhoneNumber:        req.PhoneNumber,
		Email:              req.Email,
		EmployeeID:         req.EmployeeID,
		GroupID:            req.GroupID,
		RoleID:             req.RoleID,
		IsAdmin:            req.IsAdmin,
		MustChangePassword: true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to create user"})
		return
	}

	// Preload relationships for response
	database.DB.Preload("Role").Preload("Group").First(&user, user.ID)

	c.JSON(http.StatusCreated, gin.H{
		"user":              user,
		"temporary_password": plaintextPass,
	})
}

// UpdateUser handles PUT /api/users/:id
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request fields"})
		return
	}

	// Enterprise Mode check
	if config.AppConfig.EnterpriseMode {
		if req.EmployeeID == nil || *req.EmployeeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Employee ID is required in Enterprise Mode"})
			return
		}
	}

	// Validate Group exists if provided
	if req.GroupID != nil {
		var g int64
		if err := database.DB.Model(&models.Group{}).Where("id = ?", req.GroupID).Count(&g).Error; err != nil || g == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Specified Group does not exist"})
			return
		}
	}

	// Validate Role exists if provided
	if req.RoleID != nil {
		var r int64
		if err := database.DB.Model(&models.Role{}).Where("id = ?", req.RoleID).Count(&r).Error; err != nil || r == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Specified Role does not exist"})
			return
		}
	}

	// Guardrail: check if we are demoting the only remaining administrator
	if user.IsAdmin && !req.IsAdmin {
		var adminCount int64
		database.DB.Model(&models.User{}).Where("is_admin = ?", true).Count(&adminCount)
		if adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Cannot demote the only remaining administrator"})
			return
		}
	}

	user.FullName = req.FullName
	user.PhoneNumber = req.PhoneNumber
	user.Email = req.Email
	user.EmployeeID = req.EmployeeID
	user.GroupID = req.GroupID
	user.RoleID = req.RoleID
	user.IsAdmin = req.IsAdmin

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to update user"})
		return
	}

	database.DB.Preload("Role").Preload("Group").First(&user, user.ID)
	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /api/users/:id
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	// Guardrail: cannot delete self
	currentUserVal, exists := c.Get("currentUser")
	if exists {
		currUser := currentUserVal.(*models.User)
		if currUser.ID == user.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Cannot delete your own account"})
			return
		}
	}

	// Guardrail: check if we are deleting the only remaining administrator
	if user.IsAdmin {
		var adminCount int64
		database.DB.Model(&models.User{}).Where("is_admin = ?", true).Count(&adminCount)
		if adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Cannot delete the only remaining administrator"})
			return
		}
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ResetUserPassword handles POST /api/users/:id/reset-password
func ResetUserPassword(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	var req ResetPasswordRequest
	// bind JSON is optional
	_ = c.ShouldBindJSON(&req)

	plaintextPass := req.TempPassword
	if plaintextPass == "" {
		plaintextPass = generateTempPassword(12)
	}

	hashedPass, err := services.HashPassword(plaintextPass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to encrypt password"})
		return
	}

	user.PasswordHash = hashedPass
	user.MustChangePassword = true

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Password reset successful",
		"temporary_password": plaintextPass,
	})
}
