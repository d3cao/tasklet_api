package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"tasklet_api/internal/db"
	"tasklet_api/internal/models"
	"tasklet_api/internal/testutil"
)

func requisitar(t *testing.T, router http.Handler, metodo, caminho, corpo string, destino any) int {
	t.Helper()
	resposta := httptest.NewRecorder()
	router.ServeHTTP(resposta, httptest.NewRequest(metodo, caminho, strings.NewReader(corpo)))
	if destino != nil && resposta.Code < 300 {
		if err := json.NewDecoder(resposta.Body).Decode(destino); err != nil {
			t.Fatalf("Erro ao decodificar resposta de %s %s: %v", metodo, caminho, err)
		}
	}
	return resposta.Code
}

func TestConcluirTarefaRecorrenteIntegracao(t *testing.T) {
	conn := conectarBancoDeTeste(t)
	recife, _ := time.LoadLocation("America/Recife")
	agora := time.Date(2026, 10, 10, 22, 30, 0, 0, recife)
	router := configurarRotas(conn, func() time.Time { return agora })

	var tag models.Tag
	if code := requisitar(t, router, http.MethodPost, "/tags", `{"nome":"Casa"}`, &tag); code != http.StatusCreated {
		t.Fatalf("Status esperado ao criar etiqueta: 201; recebido: %d", code)
	}
	var original models.Tarefa
	corpo := `{"nome":"Regar plantas","repeticao":7,"dia_execucao":"2026-10-01T09:00:00-03:00","prazo":"2026-10-01T18:00:00-03:00","tag_ids":[` +
		strconv.Itoa(tag.ID) + `]}`
	if code := requisitar(t, router, http.MethodPost, "/tarefas", corpo, &original); code != http.StatusCreated {
		t.Fatalf("Status esperado ao criar tarefa: 201; recebido: %d", code)
	}

	var concluida models.Tarefa
	if code := requisitar(t, router, http.MethodPost, "/tarefas/"+strconv.Itoa(original.ID)+"/concluir", "", &concluida); code != http.StatusOK {
		t.Fatalf("Status esperado ao concluir: 200; recebido: %d", code)
	}
	if concluida.Estado != models.EstadoConcluida {
		t.Errorf("Estado esperado: %d; recebido: %d", models.EstadoConcluida, concluida.Estado)
	}

	var pendentes []models.Tarefa
	requisitar(t, router, http.MethodGet, "/tarefas", "", &pendentes)
	if len(pendentes) != 1 {
		t.Fatalf("Esperada 1 próxima ocorrência pendente; recebidas %d", len(pendentes))
	}
	proxima := pendentes[0]
	diaEsperado := time.Date(2026, 10, 17, 9, 0, 0, 0, recife)
	prazoEsperado := time.Date(2026, 10, 17, 18, 0, 0, 0, recife)
	if proxima.DiaExecucao == nil || !proxima.DiaExecucao.Equal(diaEsperado) {
		t.Errorf("dia_execucao esperado: %s; recebido: %v", diaEsperado, proxima.DiaExecucao)
	}
	if proxima.Prazo == nil || !proxima.Prazo.Equal(prazoEsperado) {
		t.Errorf("prazo esperado: %s; recebido: %v", prazoEsperado, proxima.Prazo)
	}
	if proxima.Estado != models.EstadoPendente {
		t.Errorf("Estado esperado: %d; recebido: %d", models.EstadoPendente, proxima.Estado)
	}
	if len(proxima.Tags) != 1 || proxima.Tags[0].ID != tag.ID {
		t.Errorf("A próxima ocorrência deve herdar a etiqueta %d; recebidas: %v", tag.ID, proxima.Tags)
	}

	if code := requisitar(t, router, http.MethodPost, "/tarefas/"+strconv.Itoa(original.ID)+"/concluir", "", nil); code != http.StatusOK {
		t.Fatalf("Concluir de novo deve ser idempotente; status: %d", code)
	}
	requisitar(t, router, http.MethodGet, "/tarefas", "", &pendentes)
	if len(pendentes) != 1 {
		t.Errorf("Concluir de novo não pode gerar outra ocorrência; pendentes: %d", len(pendentes))
	}
}

func TestEtiquetaNomeUnicoIntegracao(t *testing.T) {
	router := ConfigurarRotas(conectarBancoDeTeste(t))

	var casa, trabalho models.Tag
	requisitar(t, router, http.MethodPost, "/tags", `{"nome":"Casa"}`, &casa)
	requisitar(t, router, http.MethodPost, "/tags", `{"nome":"Trabalho"}`, &trabalho)

	casos := []struct {
		nome, metodo, caminho, corpo string
		status                       int
	}{
		{"criar com mesmo nome", http.MethodPost, "/tags", `{"nome":"Casa"}`, http.StatusConflict},
		{"criar com maiúsculas e espaços diferentes", http.MethodPost, "/tags", `{"nome":"  cASA "}`, http.StatusConflict},
		{"renomear para nome de outra etiqueta", http.MethodPatch, "/tags/" + strconv.Itoa(trabalho.ID), `{"nome":"casa"}`, http.StatusConflict},
		{"renomear a própria etiqueta mudando só maiúsculas", http.MethodPatch, "/tags/" + strconv.Itoa(casa.ID), `{"nome":"CASA"}`, http.StatusOK},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if code := requisitar(t, router, c.metodo, c.caminho, c.corpo, nil); code != c.status {
				t.Errorf("Status esperado: %d; recebido: %d", c.status, code)
			}
		})
	}
}

func TestMigracaoFundeEtiquetasDuplicadas(t *testing.T) {
	conn := testutil.Database(t, false)
	ctx := context.Background()
	if _, err := conn.ExecContext(ctx, `CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	for _, nome := range []string{"001_base.sql", "002_recorrencia.sql"} {
		sql, err := os.ReadFile("db/migrations/" + nome)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = conn.ExecContext(ctx, string(sql)); err != nil {
			t.Fatalf("migração %s: %v", nome, err)
		}
		if _, err = conn.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, "migrations/"+nome); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO tag(tag_id, nome) VALUES (1,'Casa'), (2,'casa '), (3,'Lazer');
		INSERT INTO task(task_id, nome) VALUES (10,'A'), (11,'B');
		INSERT INTO possui_uma(task_id, tag_id) VALUES (10,1), (10,2), (11,2), (11,3);`); err != nil {
		t.Fatal(err)
	}

	if err := db.Migrate(ctx, conn); err != nil {
		t.Fatalf("Migração falhou: %v", err)
	}

	var tags int
	conn.QueryRowContext(ctx, `SELECT count(*) FROM tag`).Scan(&tags)
	if tags != 2 {
		t.Errorf("Esperadas 2 etiquetas após fundir duplicatas; recebidas %d", tags)
	}
	rows, err := conn.QueryContext(ctx, `SELECT task_id, tag_id FROM possui_uma ORDER BY task_id, tag_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var vinculos []string
	for rows.Next() {
		var task, tag int
		rows.Scan(&task, &tag)
		vinculos = append(vinculos, strconv.Itoa(task)+"-"+strconv.Itoa(tag))
	}
	if got := strings.Join(vinculos, ","); got != "10-1,11-1,11-3" {
		t.Errorf("Vínculos esperados: 10-1,11-1,11-3; recebidos: %s", got)
	}
}
