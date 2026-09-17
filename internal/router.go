package internal

import (
	"net/http"
	"tasklet_api/internal/handlers"

	"database/sql"
)

func ConfigurarRotas(conn *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	tarefaHandler := &handlers.TarefaHandler{DB: conn}

	mux.HandleFunc("GET /tarefas", tarefaHandler.ListarTarefasHandlers)
	mux.HandleFunc("POST /tarefas", tarefaHandler.CriarTarefaHandler)
	return mux
}