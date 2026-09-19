package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"tasklet_api/internal/models"
	"testing"
)

func TestListarTarefasPendentesIntegracao(t *testing.T) {
	conn := conectarBancoDeTeste(t)
	router := ConfigurarRotas(conn)

	consultar := func() []models.Tarefa {
		t.Helper()
		requisicao := httptest.NewRequest(http.MethodGet, "/tarefas", nil)
		resposta := httptest.NewRecorder()
		router.ServeHTTP(resposta, requisicao)

		if resposta.Code != http.StatusOK {
			t.Fatalf("Status esperado: %d; recebido: %d; corpo: %s", http.StatusOK, resposta.Code, resposta.Body.String())
		}
		if recebido := resposta.Header().Get("Content-Type"); recebido != "application/json" {
			t.Errorf("Content-Type esperado: application/json; recebido: %q", recebido)
		}
		if resposta.Body.String() == "null\n" {
			t.Fatal("A resposta deve ser uma lista JSON, não null")
		}

		var tarefas []models.Tarefa
		if err := json.NewDecoder(resposta.Body).Decode(&tarefas); err != nil {
			t.Fatalf("Erro ao decodificar lista de tarefas: %v", err)
		}
		return tarefas
	}

	if tarefas := consultar(); len(tarefas) != 0 {
		t.Fatalf("Lista inicial esperada vazia; recebidas %d tarefas", len(tarefas))
	}

	inserir := func(nome string, estado models.EstadoTarefa) int {
		t.Helper()
		var id int
		err := conn.QueryRowContext(context.Background(),
			"INSERT INTO task (nome, estado) VALUES ($1, $2) RETURNING task_id",
			nome, estado,
		).Scan(&id)
		if err != nil {
			t.Fatalf("Erro ao preparar tarefa de teste: %v", err)
		}
		t.Cleanup(func() {
			if _, err := conn.ExecContext(context.Background(), "DELETE FROM task WHERE task_id = $1", id); err != nil {
				t.Errorf("Erro ao remover tarefa de teste %d: %v", id, err)
			}
		})
		return id
	}

	primeiroID := inserir("Estudar Go", models.EstadoPendente)
	segundoID := inserir("Escrever testes", models.EstadoPendente)
	inserir("Tarefa concluída", models.EstadoConcluida)
	inserir("Tarefa atrasada", models.EstadoAtrasada)

	tarefas := consultar()
	if len(tarefas) != 2 {
		t.Fatalf("Esperadas 2 tarefas pendentes; recebidas %d: %+v", len(tarefas), tarefas)
	}
	if tarefas[0].ID != primeiroID || tarefas[0].Nome != "Estudar Go" || tarefas[0].Estado != models.EstadoPendente {
		t.Errorf("Primeira tarefa pendente incorreta: %+v", tarefas[0])
	}
	if tarefas[1].ID != segundoID || tarefas[1].Nome != "Escrever testes" || tarefas[1].Estado != models.EstadoPendente {
		t.Errorf("Segunda tarefa pendente incorreta: %+v", tarefas[1])
	}
}
