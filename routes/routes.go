package routes

import (
	"net/http"

	"github.com/ShreyasDr71/GoPAMS/handlers"
	"github.com/ShreyasDr71/GoPAMS/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up Gin router with auth middleware and endpoints
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Enable CORS middleware for easy frontend integration
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Serve static files for frontend index
	r.StaticFile("/", "./frontend/index.html")
	// If there are other static resources like CSS/JS
	r.Static("/assets", "./frontend/assets")

	// API Group
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			// Public routes
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", handlers.Logout)

			// Protected routes
			protected := auth.Group("")
			protected.Use(middleware.AuthRequired())
			{
				protected.POST("/change-password", handlers.ChangePassword)
				protected.GET("/me", handlers.GetMe)
			}
		}

		// Organization routes (Groups, Roles, Users)
		org := api.Group("")
		org.Use(middleware.AuthRequired())
		{
			// Groups endpoints
			org.GET("/groups", handlers.ListGroups)
			org.GET("/groups/:id", handlers.GetGroup)
			
			groupsAdmin := org.Group("/groups")
			groupsAdmin.Use(middleware.AdminRequired())
			{
				groupsAdmin.POST("", handlers.CreateGroup)
				groupsAdmin.PUT("/:id", handlers.UpdateGroup)
				groupsAdmin.DELETE("/:id", handlers.DeleteGroup)
			}

			// Roles endpoints
			org.GET("/roles", handlers.ListRoles)
			org.GET("/roles/:id", handlers.GetRole)

			rolesAdmin := org.Group("/roles")
			rolesAdmin.Use(middleware.AdminRequired())
			{
				rolesAdmin.POST("", handlers.CreateRole)
				rolesAdmin.PUT("/:id", handlers.UpdateRole)
				rolesAdmin.DELETE("/:id", handlers.DeleteRole)
			}

			// Users endpoints (restricted to Admins only)
			usersAdmin := org.Group("/users")
			usersAdmin.Use(middleware.AdminRequired())
			{
				usersAdmin.GET("", handlers.ListUsers)
				usersAdmin.GET("/:id", handlers.GetUser)
				usersAdmin.POST("", handlers.CreateUser)
				usersAdmin.PUT("/:id", handlers.UpdateUser)
				usersAdmin.DELETE("/:id", handlers.DeleteUser)
				usersAdmin.POST("/:id/reset-password", handlers.ResetUserPassword)
			}
		}
	}

	return r
}
