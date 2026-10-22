// Package testutil fornece bancos isolados para testes de integração.
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"tasklet_api/internal/db"
)

var sequence atomic.Uint64

// Database cria um schema exclusivo; nunca limpa tabelas compartilhadas.
func Database(t *testing.T, migrate bool) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL não configurada")
	}
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	admin := sql.OpenDB(stdlib.GetConnector(*config))
	t.Cleanup(func() { admin.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = admin.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), sequence.Add(1))
	if _, err = admin.ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(ctx, `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Errorf("limpar schema: %v", err)
		}
	})
	config = config.Copy()
	config.RuntimeParams["search_path"] = schema
	conn := sql.OpenDB(stdlib.GetConnector(*config))
	t.Cleanup(func() { conn.Close() })
	if migrate {
		if err = db.Migrate(ctx, conn); err != nil {
			t.Fatal(err)
		}
	}
	return conn
}
