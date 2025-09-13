package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/shaik-aaron/fetch-spotify-backend/internal/app"
	"github.com/shaik-aaron/fetch-spotify-backend/internal/database"
	"github.com/shaik-aaron/fetch-spotify-backend/internal/routes"
	_ "modernc.org/sqlite"
)

func main() {
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

	log.Println("🚀 Starting server on localhost:8080...")
	if err := router.Run("localhost:8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
