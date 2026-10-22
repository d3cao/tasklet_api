package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"tasklet_api/internal/models"
	"testing"

	"tasklet_api/internal/testutil"
)

func TestCriarTarefaIntegracao(t *testing.T) {
	conn := conectarBancoDeTeste(t)

	requisicao := httptest.NewRequest(
		http.MethodPost,
		"/tarefas",
		strings.NewReader(`{"nome":"Estudar Go"}`),
	)

	resposta := httptest.NewRecorder()

	ConfigurarRotas(conn).ServeHTTP(resposta, requisicao)

	if resposta.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar a tarefa no banco de dados: %s", resposta.Body.String())
	}

	tarefa := &models.Tarefa{}

	if err := json.NewDecoder(resposta.Body).Decode(tarefa); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	if tarefa.ID <= 0 {
		t.Fatalf("O id de uma tarefa precisa ser maior do que 0, id entregue: %v", tarefa.ID)
	}
	t.Cleanup(func() {
		if _, err := conn.ExecContext(context.Background(), "DELETE FROM task WHERE task_id = $1", tarefa.ID); err != nil {
			t.Errorf("Erro ao remover tarefa de teste: %v", err)
		}
	})

	if tarefa.Nome != "Estudar Go" {
		t.Errorf("O nome da tarefa não está correto")
	}

	if tarefa.Estado != models.EstadoPendente {
		t.Errorf("O estado esperado é: %d, estado entregue: %d", models.EstadoPendente, tarefa.Estado)
	}

	if tarefa.Repeticao != 0 {
		t.Errorf("Tarefa sem repetição definida, repetição esperada: 0, repetição entregue: %d", tarefa.Repeticao)
	}

	var nomeSalvo string
	var estadoSalvo models.EstadoTarefa
	var repeticaoSalva int

	err := conn.QueryRowContext(
		context.Background(),
		`SELECT nome, estado, repeticao
		FROM task
		WHERE task_id = $1`,
		tarefa.ID,
	).Scan(&nomeSalvo, &estadoSalvo, &repeticaoSalva)

	if err != nil {
		t.Fatalf("Erro ao consultar tarefa criada: %v", err)
	}

	if nomeSalvo != "Estudar Go" {
		t.Errorf("Nome esperado: %q; recebido: %q", "Estudar Go", nomeSalvo)
	}

	if estadoSalvo != models.EstadoPendente {
		t.Errorf("Estado esperado: %d; recebido: %d", models.EstadoPendente, estadoSalvo)
	}

	if repeticaoSalva != 0 {
		t.Errorf("Repetição esperada: 0; recebida: %d", repeticaoSalva)
	}
}

func conectarBancoDeTeste(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.Database(t, true)
}
