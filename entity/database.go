package entity

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"github.com/dgf/go-ssr-x/log"
	"github.com/google/uuid"
)

type database struct {
	db *sql.DB
}

func NewDatabase(connStr string) Storage {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("open database connection failed", err)
		os.Exit(101)
	}
	return &database{db: db}
}

func (d *database) AddTask(ctx context.Context, data TaskData) (uuid.UUID, error) {
	id := uuid.New()
	if _, err := d.db.ExecContext(ctx,
		"INSERT INTO task (id, due_date, subject, description) VALUES ($1, $2, $3, $4)",
		id, data.DueDate, data.Subject, data.Description); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (d *database) TaskCount(ctx context.Context) (int, error) {
	var count int
	if err := d.db.QueryRowContext(ctx, "SELECT count(*) FROM task").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *database) DeleteTask(ctx context.Context, id uuid.UUID) error {
	if _, err := d.db.ExecContext(ctx, "DELETE FROM task WHERE id = $1", id); err != nil {
		return err
	}
	return nil
}

func (d *database) Task(ctx context.Context, id uuid.UUID) (Task, bool, error) {
	rows := d.db.QueryRowContext(ctx,
		"SELECT id, created_at, due_date, subject, description FROM task WHERE id = $1", id)

	var task Task
	if err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject, &task.Description); err != nil {
		if err == sql.ErrNoRows {
			return task, false, nil
		}
		return task, false, err
	}
	return task, true, nil
}

func (d *database) Tasks(ctx context.Context, query TaskQuery) (TaskPage, error) {
	subjectLike := likeArg(query.Filter)
	sortOrder := taskOrderClause(query.Sort, query.Order)
	resultsQuery := "SELECT count(*) FROM task WHERE subject LIKE $1"
	rowsQuery := fmt.Sprintf("%s ORDER BY %s LIMIT %d OFFSET %d",
		"SELECT id, created_at, due_date, subject FROM task WHERE subject LIKE $1",
		sortOrder, query.Size, (query.Page-1)*query.Size)

	if tx, err := d.db.BeginTx(ctx, nil); err != nil {
		return TaskPage{}, err
	} else {
		defer tx.Rollback()
	}

	var results int
	if count, err := d.TaskCount(ctx); err != nil {
		return TaskPage{}, err
	} else if err := d.db.QueryRowContext(ctx, resultsQuery, subjectLike).Scan(&results); err != nil {
		return TaskPage{Count: count}, err
	} else if rows, err := d.db.QueryContext(ctx, rowsQuery, subjectLike); err != nil {
		return TaskPage{Count: count, Results: results}, err
	} else {
		tasks := []TaskOverview{}
		for rows.Next() {
			var task TaskOverview
			if err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject); err != nil {
				return TaskPage{Count: count, Results: results}, err
			}
			tasks = append(tasks, task)
		}
		return TaskPage{Count: count, Results: results, Tasks: tasks}, err
	}
}

func (d *database) UpdateTask(ctx context.Context, id uuid.UUID, data TaskData) (Task, bool, error) {
	if _, err := d.db.ExecContext(ctx,
		"UPDATE task SET (due_date, subject, description) = ($2, $3, $4) WHERE id = $1",
		id, data.DueDate, data.Subject, data.Description); err != nil {
		return Task{}, false, err
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
