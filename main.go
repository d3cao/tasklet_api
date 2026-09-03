package main

import (
	"fmt"
	"net/http"
	"os"

	"tasklet_api/internal"
	"tasklet_api/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {return}

	if err := db.ConnectDatabase(); err != nil {
		return
	}
	defer db.ConnectionDB.Close()

	porta := os.Getenv("PORT")
	router := internal.ConfigurarRotas()

	if err := http.ListenAndServe(porta, router); err != nil {
		fmt.Println(err)
	}

	
}