package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"

	"tasklet_api/internal/models"
	"tasklet_api/internal/repository"
)

// Field distingue ausência, null e um valor concreto em PATCH.
type Field[T any] struct {
	Set, Null bool
	Value     T
}

func (f *Field[T]) UnmarshalJSON(b []byte) error {
	f.Set = true
	f.Null = string(b) == "null"
	if f.Null {
		return nil
	}
	return json.Unmarshal(b, &f.Value)
}

type Input struct {
	Nome        Field[string]    `json:"nome"`
	Prazo       Field[time.Time] `json:"prazo"`
	Descricao   Field[string]    `json:"descricao"`
	Repeticao   Field[int]       `json:"repeticao"`
	DiaExecucao Field[time.Time] `json:"dia_execucao"`
	TagIDs      Field[[]int]     `json:"tag_ids"`
}

type Tarefas struct {
	Store *repository.Store
	Now   func() time.Time
}

func invalid(message string) error { return fmt.Errorf("%w: %s", repository.ErrInvalid, message) }

func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 50 {
		return "", invalid("nome deve ter de 1 a 50 caracteres")
	}
	return name, nil
}

func apply(t *models.Tarefa, in Input) error {
	if in.Nome.Null || in.Repeticao.Null || in.TagIDs.Null {
		return invalid("nome, repeticao e tag_ids não aceitam null")
	}
	if in.Nome.Set {
		t.Nome = in.Nome.Value
	}
	if in.Repeticao.Set {
		t.Repeticao = in.Repeticao.Value
	}
	if in.Descricao.Set {
		t.Descricao = nil
		if !in.Descricao.Null {
			v := in.Descricao.Value
			t.Descricao = &v
		}
	}
	if in.Prazo.Set {
		t.Prazo = nil
		if !in.Prazo.Null {
			v := in.Prazo.Value
			t.Prazo = &v
		}
	}
	if in.DiaExecucao.Set {
		t.DiaExecucao = nil
		if !in.DiaExecucao.Null {
			v := in.DiaExecucao.Value
			t.DiaExecucao = &v
		}
	}
	if in.TagIDs.Set {
		ids := append([]int{}, in.TagIDs.Value...)
		sort.Ints(ids)
		t.Tags = []models.Tag{}
		for i, id := range ids {
			if id <= 0 || (i > 0 && id == ids[i-1]) {
				return invalid("tag_ids deve conter IDs positivos sem duplicação")
			}
			t.Tags = append(t.Tags, models.Tag{ID: id})
		}
	}
	var err error
	t.Nome, err = ValidateName(t.Nome)
	if err != nil {
		return err
	}
	return validateDates(*t)
}

func validateDates(t models.Tarefa) error {
	if t.Repeticao < 0 {
		return invalid("repeticao deve ser maior ou igual a zero")
	}
	for _, v := range []*time.Time{t.Prazo, t.DiaExecucao} {
		if v != nil && (v.Year() < 1 || v.Year() > 9999) {
			return invalid("data fora do intervalo de 1 a 9999")
		}
	}
	if t.Prazo != nil && t.DiaExecucao != nil && t.Prazo.Before(*t.DiaExecucao) {
		return invalid("prazo não pode anteceder dia_execucao")
	}
	if t.Repeticao > 0 {
		if t.DiaExecucao == nil {
			return invalid("dia_execucao é obrigatório para tarefas recorrentes")
		}
		if _, err := advance(t.DiaExecucao, t.Repeticao); err != nil {
			return err
		}
	}
	return nil
}

var recife = func() *time.Location {
	loc, err := time.LoadLocation("America/Recife")
	if err != nil {
		panic(err)
	}
	return loc
}()

func advance(t *time.Time, days int) (*time.Time, error) {
	if t == nil {
		return nil, nil
	}
	// O limite também impede overflow em AddDate para inteiros arbitrariamente grandes.
	if days > 3660000 {
		return nil, invalid("recorrência excede o intervalo de datas suportado")
	}
	next := t.In(recife).AddDate(0, 0, days)
	return &next, inRange(next)
}
func inRange(t time.Time) error {
	if t.Year() < 1 || t.Year() > 9999 {
		return invalid("recorrência excede o intervalo de datas suportado")
	}
	return nil
}

// nextOccurrence parte do dia da conclusão, mantém o horário de dia_execucao
// e preserva a distância entre dia_execucao e prazo.
func nextOccurrence(t models.Tarefa, completedAt time.Time) (models.Tarefa, error) {
	base, done := t.DiaExecucao.In(recife), completedAt.In(recife)
	start := time.Date(done.Year(), done.Month(), done.Day(), base.Hour(), base.Minute(), base.Second(), base.Nanosecond(), recife)
	dia, err := advance(&start, t.Repeticao)
	if err != nil {
		return t, err
	}
	if t.Prazo != nil {
		prazo := dia.Add(t.Prazo.Sub(*t.DiaExecucao))
		if err = inRange(prazo); err != nil {
			return t, err
		}
		t.Prazo = &prazo
	}
	t.DiaExecucao = dia
	return t, nil
}

func effective(t models.Tarefa, now time.Time) models.Tarefa {
	if t.Estado != models.EstadoConcluida {
		t.Estado = models.EstadoPendente
		if t.Prazo != nil && t.Prazo.Before(now) {
			t.Estado = models.EstadoAtrasada
		}
	}
	return t
}
func (s *Tarefas) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
func (s *Tarefas) Create(ctx context.Context, in Input) (models.Tarefa, error) {
	t := models.Tarefa{Tags: []models.Tag{}}
	if err := apply(&t, in); err != nil {
		return t, err
	}
	t, err := s.Store.Create(ctx, t)
	return effective(t, s.now()), err
}
func (s *Tarefas) Get(ctx context.Context, id int) (models.Tarefa, error) {
	t, err := s.Store.Get(ctx, id)
	return effective(t, s.now()), err
}
func (s *Tarefas) List(ctx context.Context, f repository.Filter) ([]models.Tarefa, error) {
	now := s.now()
	tasks, err := s.Store.List(ctx, f, now)
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		tasks[i] = effective(tasks[i], now)
	}
	return tasks, nil
}
func (s *Tarefas) Patch(ctx context.Context, id int, in Input) (models.Tarefa, error) {
	t, err := s.Store.Change(ctx, id, func(t *models.Tarefa) (*models.Tarefa, error) {
		if t.Estado == models.EstadoConcluida {
			return nil, repository.ErrConflict
		}
		t.Estado = models.EstadoPendente
		return nil, apply(t, in)
	})
	return effective(t, s.now()), err
}
func (s *Tarefas) Complete(ctx context.Context, id int) (models.Tarefa, error) {
	return s.Store.Change(ctx, id, func(t *models.Tarefa) (*models.Tarefa, error) {
		if t.Estado == models.EstadoConcluida {
			return nil, nil
		}
		if err := validateDates(*t); err != nil {
			return nil, err
		}
		t.Estado = models.EstadoConcluida
		if t.Repeticao == 0 {
			return nil, nil
		}
		next, err := nextOccurrence(*t, s.now())
		if err != nil {
			return nil, err
		}
		next.ID = 0
		next.Estado = models.EstadoPendente
		origin := t.ID
		next.OrigemID = &origin
		return &next, nil
	})
}
