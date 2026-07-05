package handlers

import (
	"net/http"

	"github.com/ShreyasDr71/GoPAMS/database"
	"github.com/ShreyasDr71/GoPAMS/models"
	"github.com/gin-gonic/gin"
)

type GroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ListGroups handles GET /api/groups
func ListGroups(c *gin.Context) {
	var groups []models.Group
	if err := database.DB.Order("id asc").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to retrieve groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// GetGroup handles GET /api/groups/:id
func GetGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.Group
	if err := database.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

// CreateGroup handles POST /api/groups
func CreateGroup(c *gin.Context) {
	var req GroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Group name is required"})
		return
	}

	// Check if group name already exists
	var count int64
	database.DB.Model(&models.Group{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Group name already exists"})
		return
	}

	group := models.Group{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// UpdateGroup handles PUT /api/groups/:id
func UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.Group
	if err := database.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Group not found"})
		return
	}

	var req GroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Group name is required"})
		return
	}

	// Check for unique name conflict
	if req.Name != group.Name {
		var count int64
		database.DB.Model(&models.Group{}).Where("name = ? AND id != ?", req.Name, id).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Group name already exists"})
			return
		}
	}

	group.Name = req.Name
	group.Description = req.Description

	if err := database.DB.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup handles DELETE /api/groups/:id
func DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.Group
	if err := database.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Group not found"})
		return
	}

	// Prevent deletion of group if users are assigned to it
	var userCount int64
	database.DB.Model(&models.User{}).Where("group_id = ?", id).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Cannot delete group that has assigned users"})
		return
	}

	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}
