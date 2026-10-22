package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"tasklet_api/internal/models"
)

var (
	ErrNotFound = errors.New("recurso não encontrado")
	ErrInvalid  = errors.New("dados inválidos")
	ErrConflict = errors.New("operação incompatível com o estado atual")
	ErrTagName  = fmt.Errorf("%w: já existe etiqueta com esse nome", ErrConflict)
)

type Store struct{ DB *sql.DB }

type database interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Filter struct {
	Estado        *models.EstadoTarefa
	Todos         bool
	TagID         *int
	Limit, Offset int
}

const columns = `task_id, nome, descricao, estado, prazo, repeticao, dia_execucao, origem_id`

func normalize(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23503":
			return fmt.Errorf("%w: etiqueta inexistente", ErrInvalid)
		case "23505":
			return ErrConflict
		}
	}
	return err
}

type scanner interface{ Scan(...any) error }

func scanTask(row scanner) (models.Tarefa, error) {
	t := models.Tarefa{Tags: []models.Tag{}}
	err := row.Scan(&t.ID, &t.Nome, &t.Descricao, &t.Estado, &t.Prazo, &t.Repeticao, &t.DiaExecucao, &t.OrigemID)
	return t, normalize(err)
}

func loadTags(ctx context.Context, q database, t *models.Tarefa) error {
	rows, err := q.QueryContext(ctx, `SELECT tag.tag_id, tag.nome FROM tag JOIN possui_uma p USING(tag_id) WHERE p.task_id=$1 ORDER BY tag.tag_id`, t.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tag models.Tag
		if err = rows.Scan(&tag.ID, &tag.Nome); err != nil {
			return err
		}
		t.Tags = append(t.Tags, tag)
	}
	return rows.Err()
}

func (s *Store) Get(ctx context.Context, id int) (models.Tarefa, error) {
	t, err := scanTask(s.DB.QueryRowContext(ctx, `SELECT `+columns+` FROM task WHERE task_id=$1`, id))
	if err != nil {
		return t, err
	}
	err = loadTags(ctx, s.DB, &t)
	return t, err
}

func (s *Store) List(ctx context.Context, f Filter, now time.Time) ([]models.Tarefa, error) {
	// O mesmo instante é usado em todas as linhas e no filtro de estado.
	effective := `CASE WHEN estado=1 THEN 1 WHEN prazo < $1 THEN 2 ELSE 0 END`
	where := ` WHERE (` + effective + `) <> 1`
	args := []any{now}
	if f.Todos {
		where = ` WHERE $1::timestamptz IS NOT NULL`
	}
	if f.Estado != nil {
		args = append(args, *f.Estado)
		where = fmt.Sprintf(` WHERE (%s)=$%d`, effective, len(args))
	}
	if f.TagID != nil {
		args = append(args, *f.TagID)
		where += fmt.Sprintf(` AND EXISTS(SELECT 1 FROM possui_uma p WHERE p.task_id=task.task_id AND p.tag_id=$%d)`, len(args))
	}
	args = append(args, f.Limit, f.Offset)
	query := `SELECT ` + columns + ` FROM task` + where + fmt.Sprintf(` ORDER BY task_id LIMIT $%d OFFSET $%d`, len(args)-1, len(args))
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	tasks := []models.Tarefa{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		tasks = append(tasks, t)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		if err = loadTags(ctx, s.DB, &tasks[i]); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

func saveTags(ctx context.Context, tx *sql.Tx, t models.Tarefa) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM possui_uma WHERE task_id=$1`, t.ID); err != nil {
		return err
	}
	for _, tag := range t.Tags {
		if _, err := tx.ExecContext(ctx, `INSERT INTO possui_uma(task_id,tag_id) VALUES($1,$2)`, t.ID, tag.ID); err != nil {
			return normalize(err)
		}
	}
	return nil
}

func insertTask(ctx context.Context, tx *sql.Tx, t models.Tarefa) (models.Tarefa, error) {
	err := tx.QueryRowContext(ctx, `INSERT INTO task(nome,descricao,prazo,repeticao,dia_execucao,origem_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING task_id`, t.Nome, t.Descricao, t.Prazo, t.Repeticao, t.DiaExecucao, t.OrigemID).Scan(&t.ID)
	if err != nil {
		return t, normalize(err)
	}
	if err = saveTags(ctx, tx, t); err != nil {
		return t, err
	}
	t.Tags = []models.Tag{}
	err = loadTags(ctx, tx, &t)
	return t, err
}

func (s *Store) Create(ctx context.Context, t models.Tarefa) (models.Tarefa, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return t, err
	}
	defer tx.Rollback()
	t, err = insertTask(ctx, tx, t)
	if err != nil {
		return t, err
	}
	return t, tx.Commit()
}

// Change serializa alterações da mesma tarefa. A função pode gerar uma próxima ocorrência.
func (s *Store) Change(ctx context.Context, id int, change func(*models.Tarefa) (*models.Tarefa, error)) (models.Tarefa, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.Tarefa{}, err
	}
	defer tx.Rollback()
	t, err := scanTask(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM task WHERE task_id=$1 FOR UPDATE`, id))
	if err != nil {
		return t, err
	}
	if err = loadTags(ctx, tx, &t); err != nil {
		return t, err
	}
	next, err := change(&t)
	if err != nil {
		return t, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE task SET nome=$2,descricao=$3,estado=$4,prazo=$5,repeticao=$6,dia_execucao=$7 WHERE task_id=$1`, t.ID, t.Nome, t.Descricao, t.Estado, t.Prazo, t.Repeticao, t.DiaExecucao)
	if err != nil {
		return t, normalize(err)
	}
	if err = saveTags(ctx, tx, t); err != nil {
		return t, err
	}
	if next != nil {
		if _, err = insertTask(ctx, tx, *next); err != nil {
			return t, err
		}
	}
	t.Tags = []models.Tag{}
	if err = loadTags(ctx, tx, &t); err != nil {
		return t, err
	}
	return t, tx.Commit()
}

func (s *Store) DeleteTask(ctx context.Context, id int) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM task WHERE task_id=$1`, id)
	return deleted(result, err)
}
func deleted(result sql.Result, err error) error {
	if err != nil {
		return normalize(err)
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

func (s *Store) ListTags(ctx context.Context) ([]models.Tag, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT tag_id,nome FROM tag ORDER BY tag_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := []models.Tag{}
	for rows.Next() {
		var tag models.Tag
		if err = rows.Scan(&tag.ID, &tag.Nome); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}
func (s *Store) SaveTag(ctx context.Context, id int, name string) (models.Tag, error) {
	tag := models.Tag{Nome: name}
	query := `INSERT INTO tag(nome) VALUES($1) RETURNING tag_id`
	args := []any{name}
	if id != 0 {
		query = `UPDATE tag SET nome=$1 WHERE tag_id=$2 RETURNING tag_id`
		args = append(args, id)
	}
	err := normalize(s.DB.QueryRowContext(ctx, query, args...).Scan(&tag.ID))
	if errors.Is(err, ErrConflict) {
		err = ErrTagName
	}
	return tag, err
}
func (s *Store) DeleteTag(ctx context.Context, id int) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM tag WHERE tag_id=$1`, id)
	return deleted(result, err)
}
