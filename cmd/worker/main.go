package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("🚀 Starting Notification Worker...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ Note: No .env file found or loaded, falling back to system environment variables")
	}

	appEnv := os.Getenv("APP_ENV")
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DB_URL")
	redisURL := os.Getenv("REDIS_URL")

	log.Printf("🌍 Environment: %s", appEnv)
	log.Printf("🔌 Port: %s", port)

	// Database Connection Test
	if dbURL == "" {
		log.Println("⚠️ Warning: DB_URL is not configured")
	} else {
		log.Println("🔑 Connecting to PostgreSQL database...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		dbpool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			log.Printf("❌ Failed to create database pool: %v", err)
		} else {
			defer dbpool.Close()
			if err := dbpool.Ping(ctx); err != nil {
				log.Printf("❌ Database ping failed: %v", err)
			} else {
				log.Println("🎯 PostgreSQL connection verified successfully!")
			}
		}
	}

	// Redis Connection Indicator
	if redisURL == "" {
		log.Println("⚠️ Warning: REDIS_URL is not configured")
	} else {
		log.Printf("📡 Redis URL configured: %s", redisURL)
	}

	log.Println("🟢 Notification Worker is successfully initialized and running.")
	log.Println("ℹ️ Press Ctrl+C to terminate the process.")

	// Wait for termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("🛑 Shutting down Notification Worker gracefully...")
}