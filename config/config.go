package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	GinMode               string
	JWTSecret             string
	SessionTimeoutMinutes int
	EnterpriseMode        bool
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	DBSslMode             string
	DefaultAdminUser      string
	DefaultAdminPassword  string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file if it exists, otherwise rely on system env vars
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	AppConfig = &Config{
		Port:                 getEnv("PORT", "8080"),
		GinMode:              getEnv("GIN_MODE", "debug"),
		JWTSecret:            getEnv("JWT_SECRET", "super_secret_gopams_jwt_key_change_me_in_prod"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "gopams_user"),
		DBPassword:           getEnv("DB_PASSWORD", "gopams_password"),
		DBName:               getEnv("DB_NAME", "gopams_db"),
		DBSslMode:            getEnv("DB_SSLMODE", "disable"),
		DefaultAdminUser:     getEnv("DEFAULT_ADMIN_USER", "admin"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "AdminTempPassword123!"),
	}

	timeoutStr := getEnv("SESSION_TIMEOUT_MINUTES", "15")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = 15
	}
	AppConfig.SessionTimeoutMinutes = timeout

	enterpriseStr := getEnv("ENTERPRISE_MODE", "false")
	enterprise, err := strconv.ParseBool(enterpriseStr)
	if err != nil {
		enterprise = false
	}
	AppConfig.EnterpriseMode = enterprise
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
