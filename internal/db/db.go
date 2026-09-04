package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ConnectionDB *sql.DB

func ConnectDatabase() error {

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	port := os.Getenv("DB_PORT")
	db_name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, db_name)

	var err error

	ConnectionDB, err = sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("A conexão não foi estabelecida", "error", err)
		return err
	}

	if err = ConnectionDB.Ping(); err != nil {
		slog.Error("Não foi possível se comunicar com o banco", "error", err)
		return err
	}

	slog.Info("Conexão com o banco estabelecida")

	return nil
}
