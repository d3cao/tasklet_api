package models

import "time"

type EstadoTarefa int

//Os estados devem sempre continuar iguais, mudar eles significa uma possivel inconsistência de dados.
//Em termos de engenharia, não é a melhor forma, o ideal seria fazer código modular, porém como esse projeto foi feito apenas como um projeto para aprender Go, não tive essa preocupação
//O problema que eu estou apontando aqui é, caso o valor dos estados mude nessa constante, uma tarefa que deveria estar com um estado, estará com outro, pois o banco persistiu um valor númerico
//A conversão desse valor númerico é feita na API. Mudando a ordem dos números, muda a conversão e consequentemente os estados.
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
