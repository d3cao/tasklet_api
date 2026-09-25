package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCriarTarefaNomeVazio(t *testing.T) {
	handler := &TarefaHandler{}

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/tarefas",
		strings.NewReader(`{"nome":"    "}`),
	)
	resposta := httptest.NewRecorder()

	handler.CriarTarefaHandler(resposta, requisicao)

	if resposta.Code != http.StatusBadRequest {
		t.Errorf("Status esperado: %d; recebido %d", http.StatusBadRequest, resposta.Code)
	}
}

func TestCriarTarefaRepeticaoNegativa(t *testing.T) {
	handler := &TarefaHandler{}

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/tarefas",
		strings.NewReader(`{"nome":"Lavar os pratos", "repeticao": -1}`),
	)
	resposta := httptest.NewRecorder()

	handler.CriarTarefaHandler(resposta, requisicao)

	if resposta.Code != http.StatusBadRequest {
		t.Errorf("Status esperado: %d; recebido %d", http.StatusBadRequest, resposta.Code)
	}
}

func TestCriarTarefaJSONInvalido(t *testing.T) {
	handler := &TarefaHandler{}

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/tarefas",
		strings.NewReader(`{"nome":`),
	)
	resposta := httptest.NewRecorder()

	handler.CriarTarefaHandler(resposta, requisicao)

	if resposta.Code != http.StatusBadRequest {
		t.Errorf("Status esperado: %d; recebido %d", http.StatusBadRequest, resposta.Code)
	}
}

func TestLimiteQuantidadeCaracteres(t *testing.T) {
	handler := &TarefaHandler{}
	name := fmt.Sprintf(`{"nome":"%s"}`, strings.Repeat("a", 51))

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/tarefas",
		strings.NewReader(name),
	)
	resposta := httptest.NewRecorder()

	handler.CriarTarefaHandler(resposta, requisicao)

	if resposta.Code != http.StatusBadRequest {
		t.Errorf("Status esperado: %d; recebido %d", http.StatusBadRequest, resposta.Code)
	}
}
