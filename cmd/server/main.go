package main

import (
	"log"

	"github.com/ShreyasDr71/GoPAMS/config"
	"github.com/ShreyasDr71/GoPAMS/database"
	"github.com/ShreyasDr71/GoPAMS/routes"
)

func main() {
	log.Println("Starting GoPAMS...")

	// 1. Load Configurations
	config.LoadConfig()

	// 2. Initialize Database Connection
	database.InitDB()

	// 3. Migrate and Seed Database
	database.SeedDatabase()

	// 4. Setup Router
	r := routes.SetupRouter()

	// 5. Start Server
	port := config.AppConfig.Port
	log.Printf("GoPAMS Server is running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
