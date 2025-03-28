package sqlite3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

//go:embed *.sql
var migrations embed.FS

type file struct {
	db *sql.DB
}

func NewFile(ctx context.Context, dsn string) (entity.Storage, error) {
	if db, err := sql.Open("sqlite", dsn); err != nil {
		return nil, err
	} else if migration, err := goose.NewProvider(goose.DialectSQLite3, db, migrations); err != nil {
		return nil, err
	} else if _, err := migration.Up(ctx); err != nil {
		return nil, err
	} else {
		return &file{db: db}, nil
	}
}

func (f *file) Close() {
	if err := f.db.Close(); err != nil {
		panic(err)
	}
}

func (f *file) AddTask(ctx context.Context, data entity.TaskData) (uuid.UUID, error) {
	const query = "INSERT INTO task (id, due_date, subject, description) VALUES ($1, $2, $3, $4)"

	id := uuid.New()
	if _, err := f.db.ExecContext(ctx, query, id, data.DueDate, data.Subject, data.Description); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (f *file) DeleteTask(ctx context.Context, id uuid.UUID) error {
	const query = "DELETE FROM task WHERE id = $1"

	if result, err := f.db.ExecContext(ctx, query, id); err != nil {
		return err
	} else if rows, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("rows access failed: %w", err)
	} else if rows != 1 {
		return fmt.Errorf("no row for %s", id)
	}
	return nil
}

func (f *file) Task(ctx context.Context, id uuid.UUID) (task entity.Task, found bool, err error) {
	const query = "SELECT id, created_at, due_date, subject, description FROM task WHERE id = $1"

	row := f.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject, &task.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task, false, nil
		}
		return task, false, err
	} else {
		task.Id = id
		return task, true, nil
	}
}

func (f *file) TaskCount(ctx context.Context) (int, error) {
	const query = "SELECT count(*) FROM task"

	var count int
	if err := f.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (f *file) Tasks(ctx context.Context, query entity.TaskQuery) (entity.TaskPage, error) {
	subjectLike := likeArg(query.Filter)
	sortOrder := taskOrderClause(query.Sort, query.Order)
	resultsQuery := "SELECT count(*) FROM task WHERE subject LIKE $1"
	rowsQuery := fmt.Sprintf("%s ORDER BY %s LIMIT %d OFFSET %d",
		"SELECT id, created_at, due_date, subject FROM task WHERE subject LIKE $1",
		sortOrder, query.Size, (query.Page-1)*query.Size)

	if tx, err := f.db.BeginTx(ctx, nil); err != nil {
		return entity.TaskPage{}, err
	} else {
		defer tx.Rollback()
	}

	var results int
	if count, err := f.TaskCount(ctx); err != nil {
		return entity.TaskPage{}, err
	} else if err := f.db.QueryRowContext(ctx, resultsQuery, subjectLike).Scan(&results); err != nil {
		return entity.TaskPage{Count: count}, err
	} else if rows, err := f.db.QueryContext(ctx, rowsQuery, subjectLike); err != nil {
		return entity.TaskPage{Count: count, Results: results}, err
	} else if tasks, err := scanTaskOverviews(rows); err != nil {
		return entity.TaskPage{Count: count, Results: results}, err
	} else {
		return entity.TaskPage{Count: count, Results: results, Tasks: tasks}, nil
	}
}

func (f *file) UpdateTask(ctx context.Context, id uuid.UUID, data entity.TaskData) (task entity.Task, found bool, err error) {
	const query = "UPDATE task SET (due_date, subject, description) = ($2, $3, $4) WHERE id = $1"

	if _, err := f.db.ExecContext(ctx, query, id, data.DueDate, data.Subject, data.Description); err != nil {
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
		if err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject); err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
