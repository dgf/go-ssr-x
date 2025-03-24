package entity

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type database struct {
	db *pgxpool.Pool
}

func NewDatabase(ctx context.Context, connStr string) (Storage, error) {
	if dbpool, err := pgxpool.New(ctx, connStr); err != nil {
		return nil, err
	} else {
		return &database{db: dbpool}, nil
	}
}

func (d *database) Close() {
	d.db.Close()
}

func (d *database) AddTask(ctx context.Context, data TaskData) (uuid.UUID, error) {
	sql := "INSERT INTO task (id, due_date, subject, description) VALUES ($1, $2, $3, $4)"

	id := uuid.New()
	if _, err := d.db.Exec(ctx, sql, id, data.DueDate, data.Subject, data.Description); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (d *database) TaskCount(ctx context.Context) (int, error) {
	sql := "SELECT count(*) FROM task"

	var count int
	if err := d.db.QueryRow(ctx, sql).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *database) DeleteTask(ctx context.Context, id uuid.UUID) error {
	sql := "DELETE FROM task WHERE id = $1"

	if tag, err := d.db.Exec(ctx, sql, id); err != nil {
		return err
	} else if tag.RowsAffected() != 1 {
		return fmt.Errorf("no row for %s", id)
	}
	return nil
}

func (d *database) Task(ctx context.Context, id uuid.UUID) (Task, bool, error) {
	sql := "SELECT id, created_at, due_date, subject, description FROM task WHERE id = $1"

	if rows, err := d.db.Query(ctx, sql, id); err != nil {
		return Task{}, false, err
	} else if task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return task, false, nil
		}
		return task, false, err
	} else {
		return task, true, nil
	}
}

func (d *database) Tasks(ctx context.Context, query TaskQuery) (TaskPage, error) {
	resultsQuery := "SELECT count(*) FROM task WHERE subject LIKE $1"
	rowsQuery := fmt.Sprintf("%s ORDER BY %s LIMIT %d OFFSET %d",
		"SELECT id, created_at, due_date, subject FROM task WHERE subject LIKE $1",
		taskOrderClause(query.Sort, query.Order), query.Size, (query.Page-1)*query.Size)
	subjectLike := likeArg(query.Filter)

	if tx, err := d.db.Begin(ctx); err != nil {
		return TaskPage{}, err
	} else {
		defer tx.Rollback(ctx)
	}

	var results int
	if count, err := d.TaskCount(ctx); err != nil {
		return TaskPage{}, err
	} else if err := d.db.QueryRow(ctx, resultsQuery, subjectLike).Scan(&results); err != nil {
		return TaskPage{Count: count}, err
	} else if rows, err := d.db.Query(ctx, rowsQuery, subjectLike); err != nil {
		return TaskPage{Count: count, Results: results}, err
	} else if tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[TaskOverview]); err != nil {
		return TaskPage{Count: count, Results: results}, err
	} else {
		return TaskPage{Count: count, Results: results, Tasks: tasks}, err
	}
}

func (d *database) UpdateTask(ctx context.Context, id uuid.UUID, data TaskData) (Task, bool, error) {
	sql := "UPDATE task SET (due_date, subject, description) = ($2, $3, $4) WHERE id = $1"

	if tag, err := d.db.Exec(ctx, sql, id, data.DueDate, data.Subject, data.Description); err != nil {
		return Task{}, false, err
	} else if tag.RowsAffected() != 1 {
		return Task{}, false, fmt.Errorf("no row for %s", id)
	}
	return d.Task(ctx, id)
}

func likeArg(arg string) string {
	return "%" + strings.ReplaceAll(arg, "%", "") + "%"
}

func taskSort(sort TaskSort) string {
	switch sort {
	case TaskSortCreatedAt:
		return "created_at"
	case TaskSortDueDate:
		return "due_date"
	case TaskSortSubject:
		return "subject"
	}
	return "id"
}

func toSQLOrder(order SortOrder) string {
	if order == AscendingOrder {
		return "ASC"
	}
	return "DESC"
}

func taskOrderClause(sort TaskSort, order SortOrder) string {
	return fmt.Sprintf("%s %s", taskSort(sort), toSQLOrder(order))
}
