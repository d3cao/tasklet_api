package main

import (
	"fmt"
	"net/http"

	"tasklet_api/internal"
)

func main() {

	porta := ":8080"
	router := internal.ConfigurarRotas()

	if err := http.ListenAndServe(porta, router); err != nil {
		fmt.Println(err)
	}

	
}