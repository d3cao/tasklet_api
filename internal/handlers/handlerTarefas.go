package handlers

import (
	"encoding/json"
	"net/http"
)

func ListarTarefasHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resposta := map[string]string{"status": "sucesso", "mensagem": "Lista de tarefas"}
	json.NewEncoder(w).Encode(resposta)
}