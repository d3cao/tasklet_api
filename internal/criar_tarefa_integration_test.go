package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"tasklet_api/internal/models"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
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
}

func conectarBancoDeTeste(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL não configurada")
	}

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("Erro ao abrir banco de testes: %v", err)
	}
	t.Cleanup(func() {
		conn.Close()
	})

	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()

	if err := conn.PingContext(ctx); err != nil {
		t.Fatalf("Erro ao tentar conectar com o Banco: %v", err)
	}
	return conn
}
