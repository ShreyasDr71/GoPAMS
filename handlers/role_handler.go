package handlers

import (
	"net/http"

	"github.com/ShreyasDr71/GoPAMS/database"
	"github.com/ShreyasDr71/GoPAMS/models"
	"github.com/gin-gonic/gin"
)

type RoleRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	HierarchyLevel int    `json:"hierarchy_level"`
}

// ListRoles handles GET /api/roles
func ListRoles(c *gin.Context) {
	var roles []models.Role
	if err := database.DB.Order("hierarchy_level desc").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to retrieve roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// GetRole handles GET /api/roles/:id
func GetRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// CreateRole handles POST /api/roles
func CreateRole(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Role name is required"})
		return
	}

	// Check if role name already exists
	var count int64
	database.DB.Model(&models.Role{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Role name already exists"})
		return
	}

	role := models.Role{
		Name:           req.Name,
		Description:    req.Description,
		HierarchyLevel: req.HierarchyLevel,
	}

	if err := database.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// UpdateRole handles PUT /api/roles/:id
func UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Role not found"})
		return
	}

	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Role name is required"})
		return
	}

	// Rename check
	if req.Name != role.Name {
		if role.Name == "Administrator" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Cannot rename the default Administrator role"})
			return
		}
		var count int64
		database.DB.Model(&models.Role{}).Where("name = ? AND id != ?", req.Name, id).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Role name already exists"})
			return
		}
	}

	role.Name = req.Name
	role.Description = req.Description
	role.HierarchyLevel = req.HierarchyLevel

	if err := database.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// DeleteRole handles DELETE /api/roles/:id
func DeleteRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Role not found"})
		return
	}

	// Protect default Administrator role
	if role.Name == "Administrator" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Cannot delete the default Administrator role"})
		return
	}

	// Prevent deletion of role if users are assigned to it
	var userCount int64
	database.DB.Model(&models.User{}).Where("role_id = ?", id).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Cannot delete role that has assigned users"})
		return
	}

	if err := database.DB.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
