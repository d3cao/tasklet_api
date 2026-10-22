package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"tasklet_api/internal/models"
	"tasklet_api/internal/repository"
	"tasklet_api/internal/service"
)

type TarefaHandler struct {
	DB  *sql.DB
	Now func() time.Time
}

func (h *TarefaHandler) store() *repository.Store { return &repository.Store{DB: h.DB} }
func (h *TarefaHandler) tarefas() *service.Tarefas {
	return &service.Tarefas{Store: h.store(), Now: h.Now}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("Erro ao escrever resposta", "error", err)
	}
}
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"erro": message})
}
func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrInvalid):
		WriteError(w, 400, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		WriteError(w, 404, "Recurso não encontrado")
	case errors.Is(err, repository.ErrTagName):
		WriteError(w, 409, "Já existe etiqueta com esse nome")
	case errors.Is(err, repository.ErrConflict):
		WriteError(w, 409, "Operação incompatível com o estado atual")
	default:
		slog.Error("Falha na operação", "error", err)
		WriteError(w, 500, "Erro interno")
	}
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var raw json.RawMessage
	err := decoder.Decode(&raw)
	if err == nil {
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) == 0 || trimmed[0] != '{' {
			err = errors.New("esperado objeto JSON")
		} else {
			object := json.NewDecoder(bytes.NewReader(raw))
			object.DisallowUnknownFields()
			err = object.Decode(v)
		}
	}
	if err == nil {
		var extra any
		err = decoder.Decode(&extra)
		if errors.Is(err, io.EOF) {
			return true
		}
	}
	var max *http.MaxBytesError
	if errors.As(err, &max) {
		WriteError(w, 413, "Corpo excede 1 MiB")
	} else {
		WriteError(w, 400, "JSON inválido; envie um único objeto com campos conhecidos")
	}
	return false
}
func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		WriteError(w, 400, "ID deve ser um inteiro positivo")
		return 0, false
	}
	return id, true
}
func (h *TarefaHandler) CriarTarefaHandler(w http.ResponseWriter, r *http.Request) {
	var in service.Input
	if !decode(w, r, &in) {
		return
	}
	t, err := h.tarefas().Create(r.Context(), in)
	if err != nil {
		respondError(w, err)
		return
	}
	w.Header().Set("Location", "/tarefas/"+strconv.Itoa(t.ID))
	WriteJSON(w, 201, t)
}
func (h *TarefaHandler) ListarTarefasPendentesHandlers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.Filter{Limit: 50}
	for key, values := range q {
		if len(values) != 1 {
			WriteError(w, 400, "Parâmetro repetido: "+key)
			return
		}
		value := values[0]
		switch key {
		case "estado":
			if value == "todos" {
				f.Todos = true
				continue
			}
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 || n > 2 {
				WriteError(w, 400, "Estado inválido")
				return
			}
			state := models.EstadoTarefa(n)
			f.Estado = &state
		case "tag_id", "limit", "offset":
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 || (key != "offset" && n == 0) || (key == "limit" && n > 100) {
				WriteError(w, 400, "Parâmetro inválido: "+key)
				return
			}
			switch key {
			case "tag_id":
				f.TagID = &n
			case "limit":
				f.Limit = n
			case "offset":
				f.Offset = n
			}
		default:
			WriteError(w, 400, "Parâmetro desconhecido: "+key)
			return
		}
	}
	tasks, err := h.tarefas().List(r.Context(), f)
	if err != nil {
		respondError(w, err)
		return
	}
	WriteJSON(w, 200, tasks)
}
func (h *TarefaHandler) ObterTarefa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	t, err := h.tarefas().Get(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	WriteJSON(w, 200, t)
}
func (h *TarefaHandler) EditarTarefa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in service.Input
	if !decode(w, r, &in) {
		return
	}
	t, err := h.tarefas().Patch(r.Context(), id, in)
	if err != nil {
		respondError(w, err)
		return
	}
	WriteJSON(w, 200, t)
}
func (h *TarefaHandler) ConcluirTarefa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	// Endpoint sem corpo: ainda limita e rejeita conteúdo inesperado.
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		var max *http.MaxBytesError
		if errors.As(err, &max) {
			WriteError(w, 413, "Corpo excede 1 MiB")
		} else {
			WriteError(w, 400, "Corpo inválido")
		}
		return
	}
	if len(body) != 0 {
		WriteError(w, 400, "A conclusão não aceita corpo")
		return
	}
	t, err := h.tarefas().Complete(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	WriteJSON(w, 200, t)
}
func (h *TarefaHandler) ExcluirTarefa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.store().DeleteTask(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (h *TarefaHandler) ListarTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.store().ListTags(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	WriteJSON(w, 200, tags)
}
func (h *TarefaHandler) SalvarTag(w http.ResponseWriter, r *http.Request) {
	id := 0
	if r.Method == http.MethodPatch {
		var ok bool
		id, ok = pathID(w, r)
		if !ok {
			return
		}
	}
	var in struct {
		Nome string `json:"nome"`
	}
	if !decode(w, r, &in) {
		return
	}
	name, err := service.ValidateName(in.Nome)
	if err != nil {
		respondError(w, err)
		return
	}
	tag, err := h.store().SaveTag(r.Context(), id, name)
	if err != nil {
		respondError(w, err)
		return
	}
	status := 200
	if id == 0 {
		status = 201
	}
	WriteJSON(w, status, tag)
}
func (h *TarefaHandler) ExcluirTag(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.store().DeleteTag(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(204)
}
