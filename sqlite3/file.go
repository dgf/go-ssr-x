// Package sqlite3 provides a SQLite version 3 backed entity storage.
package sqlite3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/log"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrations embed.FS

type file struct {
	db *sql.DB
}

func NewFile(ctx context.Context, dsn string) (entity.Storage, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	migration, err := goose.NewProvider(goose.DialectSQLite3, db, migrations)
	if err != nil {
		return nil, err
	}

	_, err = migration.Up(ctx)
	if err != nil {
		return nil, err
	}

	return &file{db: db}, nil
}

func (f *file) Close() error {
	return f.db.Close()
}

func (f *file) AddTask(ctx context.Context, data entity.TaskData) (uuid.UUID, error) {
	const query = "INSERT INTO task (id, due_date, subject, description) VALUES ($1, $2, $3, $4)"

	id := uuid.New()
	_, err := f.db.ExecContext(ctx, query, id, data.DueDate, data.Subject, data.Description)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (f *file) DeleteTask(ctx context.Context, id uuid.UUID) error {
	const query = "DELETE FROM task WHERE id = $1"

	result, err := f.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows access failed: %w", err)
	}

	if rows != 1 {
		return fmt.Errorf("no row for %s", id)
	}

	return nil
}

func (f *file) Task(ctx context.Context, id uuid.UUID) (entity.Task, bool, error) {
	const query = "SELECT created_at, due_date, subject, description FROM task WHERE id = $1"

	var task entity.Task
	row := f.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&task.CreatedAt, &task.DueDate, &task.Subject, &task.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task, false, nil
		}

		return task, false, err
	}

	task.Id = id

	return task, true, nil
}

func (f *file) TaskCount(ctx context.Context) (int, error) {
	const query = "SELECT count(*) FROM task"

	var count int
	err := f.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (f *file) Tasks(ctx context.Context, query entity.TaskQuery) (entity.TaskPage, error) {
	const resultsQuery = "SELECT count(*) FROM task WHERE subject LIKE $1"
	const rowSelectQuery = "SELECT id, created_at, due_date, subject FROM task WHERE subject LIKE $1"
	rowsQuery := fmt.Sprintf("%s ORDER BY %s LIMIT %d OFFSET %d", rowSelectQuery,
		taskOrderClause(query.Sort, query.Order), query.Size, (query.Page-1)*query.Size)
	subjectLike := likeArg(query.Filter)

	tx, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.TaskPage{}, err
	}

	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Error("task list query read rollback failed", err)
		}
	}()

	var results int
	count, err := f.TaskCount(ctx)
	if err != nil {
		return entity.TaskPage{}, err
	}

	err = f.db.QueryRowContext(ctx, resultsQuery, subjectLike).Scan(&results)
	if err != nil {
		return entity.TaskPage{Count: count}, err
	}

	rows, err := f.db.QueryContext(ctx, rowsQuery, subjectLike)
	if err != nil {
		return entity.TaskPage{Count: count, Results: results}, err
	}

	tasks, err := scanTaskOverviews(rows)
	if err != nil {
		return entity.TaskPage{Count: count, Results: results}, err
	}

	return entity.TaskPage{Count: count, Results: results, Tasks: tasks}, nil
}

func (f *file) UpdateTask(ctx context.Context, id uuid.UUID, data entity.TaskData) (entity.Task, bool, error) {
	const query = "UPDATE task SET (due_date, subject, description) = ($2, $3, $4) WHERE id = $1"

	_, err := f.db.ExecContext(ctx, query, id, data.DueDate, data.Subject, data.Description)
	if err != nil {
		return entity.Task{}, false, err
	}

	return f.Task(ctx, id)
}

func likeArg(arg string) string {
	return "%" + strings.ReplaceAll(arg, "%", "") + "%"
}

func taskSort(sort entity.TaskSort) string {
	switch sort {
	case entity.TaskSortCreatedAt:
		return "created_at"
	case entity.TaskSortDueDate:
		return "due_date"
	case entity.TaskSortSubject:
		return "subject"
	}

	return "id"
}

func toSQLOrder(order entity.SortOrder) string {
	if order == entity.AscendingOrder {
		return "ASC"
	}

	return "DESC"
}

func taskOrderClause(sort entity.TaskSort, order entity.SortOrder) string {
	return fmt.Sprintf("%s %s", taskSort(sort), toSQLOrder(order))
}

func scanTaskOverviews(rows *sql.Rows) ([]entity.TaskOverview, error) {
	tasks := []entity.TaskOverview{}

	for rows.Next() {
		var task entity.TaskOverview
		err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject)
		if err != nil {
			return tasks, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
