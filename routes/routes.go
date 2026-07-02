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
	}

	return r
}
