package database

import (
	"fmt"
	"log"
	"time"

	"github.com/ShreyasDr71/GoPAMS/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes connection to PostgreSQL with retry logic
func InitDB() {
	var err error
	cfg := config.AppConfig

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSslMode,
	)

	maxRetries := 5
	for i := 1; i <= maxRetries; i++ {
		log.Printf("Connecting to database (attempt %d/%d)...", i, maxRetries)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Successfully connected to the database!")
			return
		}

		log.Printf("Failed to connect to database: %v. Retrying in 3 seconds...", err)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("Could not connect to database after %d attempts: %v", maxRetries, err)
}
