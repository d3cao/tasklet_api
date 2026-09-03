package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ConnectionDB *sql.DB

func ConnectDatabase() error {
	dsn := os.Getenv("DATABASE_URL")
	
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