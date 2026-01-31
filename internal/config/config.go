package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port         string
	MongoURI     string
	JWTSecret    string
	JWTExpiresIn int
}

var AppConfig *Config

// Load initializes configuration from environment variables
func Load() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGODB_ATLAS_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_ATLAS_URI environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	jwtExpiresIn := 3600 // default 1 hour
	if expStr := os.Getenv("JWT_EXPIRES_IN"); expStr != "" {
		if exp, err := strconv.Atoi(expStr); err == nil {
			jwtExpiresIn = exp
		}
	}

	AppConfig = &Config{
		Port:         port,
		MongoURI:     mongoURI,
		JWTSecret:    jwtSecret,
		JWTExpiresIn: jwtExpiresIn,
	}
}
