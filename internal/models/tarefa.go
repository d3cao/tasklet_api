package models

import "time"

type EstadoTarefa int

// Valores persistidos: não alterar a numeração.
const (
	EstadoPendente  EstadoTarefa = 0
	EstadoConcluida EstadoTarefa = 1
	EstadoAtrasada  EstadoTarefa = 2
)

type Tag struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

type Tarefa struct {
	ID          int          `json:"id"`
	Nome        string       `json:"nome"`
	Prazo       *time.Time   `json:"prazo"`
	Descricao   *string      `json:"descricao"`
	Estado      EstadoTarefa `json:"estado"`
	Repeticao   int          `json:"repeticao"`
	DiaExecucao *time.Time   `json:"dia_execucao"`
	Tags        []Tag        `json:"tags"`
	OrigemID    *int         `json:"-"`
}
