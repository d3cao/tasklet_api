package internal

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestCriarTarefaIntegracao(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL não condigurada")
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
}