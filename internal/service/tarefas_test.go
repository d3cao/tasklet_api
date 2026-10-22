package service

import (
	"testing"
	"time"

	"tasklet_api/internal/models"
)

func data(t *testing.T, valor string) *time.Time {
	t.Helper()
	d, err := time.ParseInLocation("2006-01-02 15:04", valor, recife)
	if err != nil {
		t.Fatal(err)
	}
	return &d
}

func TestProximaOcorrenciaPartirDaConclusao(t *testing.T) {
	casos := []struct {
		nome, dia, prazo, conclusao string
		conclusaoUTC                bool
		repeticao                   int
		diaEsperado, prazoEsperado  string
	}{
		{nome: "concluída com atraso", dia: "2026-10-01 09:00", prazo: "2026-10-01 18:00", conclusao: "2026-10-10 22:30",
			repeticao: 7, diaEsperado: "2026-10-17 09:00", prazoEsperado: "2026-10-17 18:00"},
		{nome: "concluída antes do dia de execução", dia: "2026-10-20 09:00", prazo: "2026-10-21 09:00", conclusao: "2026-10-15 08:00",
			repeticao: 7, diaEsperado: "2026-10-22 09:00", prazoEsperado: "2026-10-23 09:00"},
		{nome: "dia da conclusão no fuso de Recife", dia: "2026-10-01 09:00", conclusao: "2026-10-10 22:00", conclusaoUTC: true,
			repeticao: 1, diaEsperado: "2026-10-11 09:00"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			tarefa := models.Tarefa{Repeticao: c.repeticao, DiaExecucao: data(t, c.dia)}
			if c.prazo != "" {
				tarefa.Prazo = data(t, c.prazo)
			}
			conclusao := *data(t, c.conclusao)
			if c.conclusaoUTC {
				conclusao = conclusao.UTC() // 01:00 UTC do dia seguinte
			}
			proxima, err := nextOccurrence(tarefa, conclusao)
			if err != nil {
				t.Fatal(err)
			}
			if !proxima.DiaExecucao.Equal(*data(t, c.diaEsperado)) {
				t.Errorf("dia_execucao esperado: %s; recebido: %s", c.diaEsperado, proxima.DiaExecucao.In(recife))
			}
			switch {
			case c.prazoEsperado == "" && proxima.Prazo != nil:
				t.Errorf("prazo esperado nulo; recebido: %s", proxima.Prazo)
			case c.prazoEsperado != "" && (proxima.Prazo == nil || !proxima.Prazo.Equal(*data(t, c.prazoEsperado))):
				t.Errorf("prazo esperado: %s; recebido: %v", c.prazoEsperado, proxima.Prazo)
			}
		})
	}
}

func TestProximaOcorrenciaForaDoIntervalo(t *testing.T) {
	tarefa := models.Tarefa{Repeticao: 2, DiaExecucao: data(t, "9999-12-30 09:00")}
	if _, err := nextOccurrence(tarefa, *data(t, "9999-12-30 10:00")); err == nil {
		t.Fatal("esperado erro para data além do ano 9999")
	}
}
