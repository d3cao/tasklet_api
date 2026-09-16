package models

import "time"

type Tarefa struct {
	ID int `json:"id"`
	Nome string `json:"nome"`
	Prazo *time.Time `json:"prazo"`
	Descricao *string `json:"descricao"`
	Estado int `json:"estado"`
	Repeticao int `json:"repeticao"`
	DiaExecucao *time.Time `json:"dia_execucao"`
}
