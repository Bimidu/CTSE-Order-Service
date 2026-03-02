package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	GRPCPort        string
	DatabaseURL     string
	JWTSecret       string
	AuthServiceAddr string
	ProductServiceAddr string
	Env             string
}

var App *Config

func Load() {
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	App = &Config{
		Port:               getEnv("PORT", "6969"),
		GRPCPort:           getEnv("GRPC_PORT", "50052"),
		DatabaseURL:        mustGetEnv("DATABASE_URL"),
		JWTSecret:          mustGetEnv("JWT_SECRET"),
		AuthServiceAddr:    getEnv("AUTH_SERVICE_ADDR", "auth-service:50051"),
		ProductServiceAddr: getEnv("PRODUCT_SERVICE_ADDR", "product-service:50053"),
		Env:                getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return v
}
