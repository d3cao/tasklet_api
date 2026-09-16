package models

import "time"

type EstadoTarefa int

const (
	EstadoPendente EstadoTarefa = 0
	EstadoConcluida EstadoTarefa = 1
	EstadoAtrasada EstadoTarefa = 2
)

type Tarefa struct {
	ID int `json:"id"`
	Nome string `json:"nome"`
	Prazo *time.Time `json:"prazo"`
	Descricao *string `json:"descricao"`
	Estado EstadoTarefa `json:"estado"`
	Repeticao int `json:"repeticao"`
	DiaExecucao *time.Time `json:"dia_execucao"`
}
