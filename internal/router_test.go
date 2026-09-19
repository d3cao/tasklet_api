package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
