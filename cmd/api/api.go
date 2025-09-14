package main

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/shaik-aaron/fetch-spotify-backend/internal/app"
	"github.com/shaik-aaron/fetch-spotify-backend/internal/database"
	"github.com/shaik-aaron/fetch-spotify-backend/internal/routes"
	_ "modernc.org/sqlite"
)

func main() {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	dbPath := "./data/tokens.db"
	db, err := sqlx.Connect("sqlite", dbPath+"?_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Set connection pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := database.RunMigrations(db, "./migrations"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Use the App struct from internal/app
	appInstance := app.New(db)

	router := routes.SetupRoutes(appInstance)

	// Use PORT environment variable for Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
