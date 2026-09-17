package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListarTarefas(t *testing.T) {
	//Verifica se a requisição foi bem sucedida ou não
	router := ConfigurarRotas(nil)

	requisicao := httptest.NewRequest(http.MethodGet, "/tarefas", nil)
	resposta := httptest.NewRecorder()

	router.ServeHTTP(resposta, requisicao)

	if resposta.Code != http.StatusOK {
		t.Errorf("Status esperado: %d; recebido: %d", http.StatusOK, resposta.Code)
	}

	if recebido := resposta.Header().Get("Content-Type"); recebido != "application/json" {
		t.Errorf("Content-Type esperado: application/json; recebido: %q", recebido)
	}

	//Valida se o JSON está correto e verifica os campos para saber se há algum problema
	var corpo map[string]string
	
	if err := json.NewDecoder(resposta.Body).Decode(&corpo); err != nil {
		t.Fatalf("Erro ao decodificar resposta JSON: %v", err)
	}

	if recebido := corpo["status"]; recebido != "sucesso" {
		t.Errorf("status esperado: sucesso; recebido: %q", recebido)
	}
}

func TestCriarTarefaRotaRegistrada(t *testing.T) {
	router := ConfigurarRotas(nil)

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/tarefas",
		strings.NewReader(`{"nome":""}`),
	)
	resposta := httptest.NewRecorder()

	router.ServeHTTP(resposta, requisicao)

	if resposta.Code != http.StatusBadRequest {
		t.Errorf("Status esperado: %d; recebido: %d", http.StatusBadRequest, resposta.Code)
	}
}

