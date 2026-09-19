package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"tasklet_api/internal/models"
	"time"
	"unicode/utf8"
)

type TarefaHandler struct {
	DB *sql.DB
}

type CriarTarefaEntrada struct {
	Nome string `json:"nome"`
	Prazo *time.Time `json:"prazo"`
	Descricao *string `json:"descricao"`
	Repeticao int `json:"repeticao"`
	DiaExecucao *time.Time `json:"dia_execucao"`
}

const query = `
	INSERT INTO task (nome, descricao, prazo, repeticao, dia_execucao)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING task_id, nome, descricao, estado, prazo, repeticao, dia_execucao
`

func (h *TarefaHandler) CriarTarefaHandler(w http.ResponseWriter, r *http.Request) {
	var entrada CriarTarefaEntrada
	var tarefa models.Tarefa

	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	entrada.Nome = strings.TrimSpace(entrada.Nome)
	if entrada.Nome == "" {
		http.Error(w, "Nome é obrigatório", http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(entrada.Nome) > 50 {
		http.Error(w, "Nome deve possuir no máximo 50 caracteres", http.StatusBadRequest)
		return
	}

	if entrada.Repeticao < 0 {
		http.Error(w, "A repetição não pode ser menor do que 0", http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRowContext(
		r.Context(),
		query,
		entrada.Nome,
		entrada.Descricao,
		entrada.Prazo,
		entrada.Repeticao,
		entrada.DiaExecucao,
	).Scan(
		&tarefa.ID,
		&tarefa.Nome,
		&tarefa.Descricao,
		&tarefa.Estado,
		&tarefa.Prazo,
		&tarefa.Repeticao,
		&tarefa.DiaExecucao,
	)

	if err != nil {
		slog.Error("Erro ao salvar informações no banco de dados", "error", err)
		http.Error(w, "Erro ao criar tarefa", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	if err := json.NewEncoder(w).Encode(tarefa); err != nil {
		slog.Error("Erro ao escrever resposta da criação de tarefa", "error", err)
	}
}

func (h *TarefaHandler) ListarTarefasPendentesHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resposta := map[string]string{"status": "sucesso", "mensagem": "Lista de tarefas"}
	json.NewEncoder(w).Encode(resposta)
}