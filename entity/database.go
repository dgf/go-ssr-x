package entity

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

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

func (d *database) AddTask(dueDate time.Time, subject, description string) (uuid.UUID, error) {
	id := uuid.New()
	if _, err := d.db.Exec(
		"INSERT INTO task (id, due_date, subject, description) VALUES ($1, $2, $3, $4)",
		id, dueDate, subject, description); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (d *database) TaskCount() (int, error) {
	var count int
	if err := d.db.QueryRow("SELECT count(*) FROM task").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *database) DeleteTask(id uuid.UUID) error {
	if _, err := d.db.Exec("DELETE FROM task WHERE id = $1", id); err != nil {
		return err
	}
	return nil
}

func (d *database) Task(id uuid.UUID) (Task, bool, error) {
	rows := d.db.QueryRow("SELECT id, created_at, due_date, subject, description FROM task WHERE id = $1", id)

	var task Task
	if err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject, &task.Desciption); err != nil {
		if err == sql.ErrNoRows {
			return task, false, nil
		}
		return task, false, err
	}
	return task, true, nil
}

func (d *database) Tasks(page TaskPage) ([]TaskOverview, error) {
	sql := "SELECT id, created_at, due_date, subject FROM task WHERE subject LIKE $1"
	sortOrder := taskOrderClause(page.Sort, page.Order)
	query := fmt.Sprintf("%s ORDER BY %s LIMIT %d OFFSET %d",
		sql, sortOrder, page.Size, (page.Page-1)*page.Size)

	var tasks []TaskOverview
	if rows, err := d.db.Query(query, likeArg(page.Filter)); err != nil {
		return tasks, err
	} else {
		for rows.Next() {
			var task TaskOverview
			if err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject); err != nil {
				return tasks, err
			}
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (d *database) UpdateTask(id uuid.UUID, dueDate time.Time, subject, description string) (Task, bool, error) {
	if _, err := d.db.Exec(
		"UPDATE task SET (due_date, subject, description) = ($2, $3, $4) WHERE id = $1",
		id, dueDate, subject, description); err != nil {
		return Task{}, false, err
	}
	return d.Task(id)
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
