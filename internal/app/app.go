package app

import "github.com/jmoiron/sqlx"

type App struct {
	DB *sqlx.DB
}

func New(db *sqlx.DB) *App {
	return &App{DB: db}
}
