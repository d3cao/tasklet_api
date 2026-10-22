package internal

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"tasklet_api/internal/handlers"
)

func ConfigurarRotas(conn *sql.DB) *http.ServeMux { return configurarRotas(conn, time.Now) }

func configurarRotas(conn *sql.DB, now func() time.Time) *http.ServeMux {
	mux := http.NewServeMux()
	h := &handlers.TarefaHandler{DB: conn, Now: now}
	mux.HandleFunc("GET /tarefas", h.ListarTarefasPendentesHandlers)
	mux.HandleFunc("POST /tarefas", h.CriarTarefaHandler)
	mux.HandleFunc("GET /tarefas/{id}", h.ObterTarefa)
	mux.HandleFunc("PATCH /tarefas/{id}", h.EditarTarefa)
	mux.HandleFunc("DELETE /tarefas/{id}", h.ExcluirTarefa)
	mux.HandleFunc("POST /tarefas/{id}/concluir", h.ConcluirTarefa)
	mux.HandleFunc("GET /tags", h.ListarTags)
	mux.HandleFunc("POST /tags", h.SalvarTag)
	mux.HandleFunc("PATCH /tags/{id}", h.SalvarTag)
	mux.HandleFunc("DELETE /tags/{id}", h.ExcluirTag)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSON(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if conn == nil || conn.PingContext(ctx) != nil {
			handlers.WriteError(w, 503, "Banco indisponível")
			return
		}
		handlers.WriteJSON(w, 200, map[string]string{"status": "ok"})
	})
	return mux
}
