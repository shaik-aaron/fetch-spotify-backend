package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Token struct {
	ID           int    `db:"id" json:"id"`
	RefreshToken string `db:"refresh_token" json:"refresh_token"`
	AccessToken  string `db:"access_token" json:"access_token"`
	CreatedAt    string `db:"created_at" json:"created_at"`
	UpdatedAt    string `db:"updated_at" json:"updated_at"`
}

func Close(db *sqlx.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func RunMigrations(db *sqlx.DB, migrationsDir string) error {

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	if len(files) == 0 {
		log.Println("⚠️  No migration files found in", migrationsDir)
		return nil
	}

	for _, file := range files {
		log.Printf("📄 Running migration: %s", filepath.Base(file))

		sqlBytes, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	log.Println("✅ All migrations completed successfully")
	return nil
}
