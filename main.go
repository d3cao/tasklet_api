package main

import (
	"net/http"
	"os"
	"log/slog"

	"tasklet_api/internal"
	"tasklet_api/internal/db"

	"github.com/joho/godotenv"
)

func main() {

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.Info("Log iniciado com sucesso")

	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", "error", err)
	}

	if err := db.ConnectDatabase(); err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.ConnectionDB.Close()

	porta := os.Getenv("PORT")
	router := internal.ConfigurarRotas()

	if err := http.ListenAndServe(porta, router); err != nil {
		slog.Error("HTTP server error", "error", err)
	}

	
}