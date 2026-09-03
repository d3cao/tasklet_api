package db

import (
	"database/sql"
	"fmt"
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
		fmt.Println("[Erro]: A conexão não foi estabelecida")
		return err
	}

	if err = ConnectionDB.Ping(); err != nil {
		fmt.Println("[Erro]: Não foi possível se comunicar com o banco")
		return err
	}

	fmt.Println("Conexão com o banco estabelecida")

	return nil
}