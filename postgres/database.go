// Package postgres provides a PostgreSQL backed entity storage.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type database struct {
	db *pgxpool.Pool
}

func NewDatabase(ctx context.Context, connStr string) (entity.Storage, error) {
	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return &database{db: dbpool}, nil
}

func (d *database) Close() error {
	d.db.Close()

	return nil
}

func (d *database) AddTask(ctx context.Context, data entity.TaskData) (uuid.UUID, error) {
	const sql = "INSERT INTO task (id, due_date, subject, description) VALUES ($1, $2, $3, $4)"

	id := uuid.New()
	_, err := d.db.Exec(ctx, sql, id, data.DueDate, data.Subject, data.Description)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (d *database) DeleteTask(ctx context.Context, id uuid.UUID) error {
	const sql = "DELETE FROM task WHERE id = $1"

	tag, err := d.db.Exec(ctx, sql, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("no row for %s", id)
	}

	return nil
}

func (d *database) Task(ctx context.Context, id uuid.UUID) (entity.Task, bool, error) {
	const sql = "SELECT id, created_at, due_date, subject, description FROM task WHERE id = $1"

	rows, err := d.db.Query(ctx, sql, id)
	if err != nil {
		return entity.Task{}, false, err
	}

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[entity.Task])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return task, false, nil
		}

		return task, false, err
	}

	return task, true, nil
}

func (d *database) TaskCount(ctx context.Context) (int, error) {
	const sql = "SELECT count(*) FROM task"

	var count int
	err := d.db.QueryRow(ctx, sql).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (d *database) Tasks(ctx context.Context, query entity.TaskQuery) (entity.TaskPage, error) {
	const resultsQuery = "SELECT count(*) FROM task WHERE subject LIKE $1"
	const rowSelectQuery = "SELECT id, created_at, due_date, subject FROM task WHERE subject LIKE $1"
	rowsQuery := fmt.Sprintf("%s ORDER BY %s LIMIT %d OFFSET %d", rowSelectQuery,
		taskOrderClause(query.Sort, query.Order), query.Size, (query.Page-1)*query.Size)
	subjectLike := likeArg(query.Filter)

	tx, err := d.db.Begin(ctx)
	if err != nil {
		return entity.TaskPage{}, err
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Error("task list query read rollback failed", err)
		}
	}()

	var results int
	count, err := d.TaskCount(ctx)
	if err != nil {
		return entity.TaskPage{}, err
	}

	err = d.db.QueryRow(ctx, resultsQuery, subjectLike).Scan(&results)
	if err != nil {
		return entity.TaskPage{Count: count}, err
	}

	rows, err := d.db.Query(ctx, rowsQuery, subjectLike)
	if err != nil {
		return entity.TaskPage{Count: count, Results: results}, err
	}

	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.TaskOverview])
	if err != nil {
		return entity.TaskPage{Count: count, Results: results}, err
	}

	return entity.TaskPage{Count: count, Results: results, Tasks: tasks}, nil
}

func (d *database) UpdateTask(ctx context.Context, id uuid.UUID, data entity.TaskData) (entity.Task, bool, error) {
	const sql = "UPDATE task SET (due_date, subject, description) = ($2, $3, $4) WHERE id = $1"

	tag, err := d.db.Exec(ctx, sql, id, data.DueDate, data.Subject, data.Description)
	if err != nil {
		return entity.Task{}, false, err
	}

	if tag.RowsAffected() != 1 {
		return entity.Task{}, false, fmt.Errorf("no row for %s", id)
	}

	return d.Task(ctx, id)
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
