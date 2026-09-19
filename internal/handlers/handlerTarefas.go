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

const queryInsertion = `
	INSERT INTO task (nome, descricao, prazo, repeticao, dia_execucao)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING task_id, nome, descricao, estado, prazo, repeticao, dia_execucao
`

const queryPending = `
	SELECT task_id, nome, descricao, estado, prazo, repeticao, dia_execucao
	FROM task
	WHERE estado = $1
	ORDER BY task_id
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
		queryInsertion,
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

	rows, err := h.DB.QueryContext(r.Context(), queryPending, models.EstadoPendente)
	if err != nil {
		slog.Error("Erro ao fazer a query de busca no banco de dados", "error", err)
		http.Error(w, "Erro ao listar tarefas", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tarefas := []models.Tarefa{}

	for rows.Next() {
		var tarefa models.Tarefa

		if err := rows.Scan(
			&tarefa.ID,
			&tarefa.Nome,
			&tarefa.Descricao,
			&tarefa.Estado,
			&tarefa.Prazo,
			&tarefa.Repeticao,
			&tarefa.DiaExecucao,
		); err != nil {
			slog.Error("Erro no scan das tarefas (rows.Scan)", "error", err)
			http.Error(w, "Erro ao ler as tarefas", http.StatusInternalServerError)
			return
		}

		tarefas = append(tarefas, tarefa)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Erro durante a consulta das rows", "error", err)
		http.Error(w, "Erro durante a consulta dos dados", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tarefas)
}