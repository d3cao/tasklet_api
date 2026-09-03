package internal

import (
	"net/http"
	"tasklet_api/internal/handlers"
)

func ConfigurarRotas() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tarefas", handlers.ListarTarefasHandlers)
	return mux
}