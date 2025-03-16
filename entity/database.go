package entity

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/dgf/go-ssr-x/log"
	"github.com/google/uuid"
)

type database struct {
	db *sql.DB
}

const (
	insertTaskSQL = "INSERT INTO task (id, due_date, subject, description) VALUES ($1, $2, $3, $4)"
	countTasksSQL = "SELECT count(*) FROM task"
	selectTaskSQL = "SELECT id, created_at, due_date, subject, description FROM task"
	deleteTaskSQL = "DELETE FROM task WHERE id = $1"
	updateTaskSQL = "UPDATE task SET (due_date, subject, description) = ($2, $3, $4) WHERE id = $1"
)

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
	if _, err := d.db.Exec(insertTaskSQL, id, dueDate, subject, description); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (d *database) TaskCount() (int, error) {
	var count int
	r := d.db.QueryRow(countTasksSQL)
	if err := r.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *database) DeleteTask(id uuid.UUID) error {
	if _, err := d.db.Exec(deleteTaskSQL, id); err != nil {
		return err
	}
	return nil
}

func (d *database) Task(id uuid.UUID) (Task, bool, error) {
	var task Task
	rows := d.db.QueryRow(fmt.Sprintf("%s WHERE id = $1", selectTaskSQL), id)
	if err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject, &task.Desciption); err != nil {
		if err == sql.ErrNoRows {
			return task, false, nil
		}
		return task, false, err
	}
	return task, true, nil
}

func (d *database) Tasks(order TaskOrder) ([]Task, error) {
	var tasks []Task
	if rows, err := d.db.Query(fmt.Sprintf("%s ORDER BY %s", selectTaskSQL, taskOrderClause(order))); err != nil {
		return tasks, err
	} else {
		for rows.Next() {
			var task Task
			if err := rows.Scan(&task.Id, &task.CreatedAt, &task.DueDate, &task.Subject, &task.Desciption); err != nil {
				return tasks, err
			}
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (d *database) UpdateTask(id uuid.UUID, dueDate time.Time, subject, description string) (Task, bool, error) {
	if _, err := d.db.Exec(updateTaskSQL, id, dueDate, subject, description); err != nil {
		return Task{}, false, err
	}
	return d.Task(id)
}

func taskOrderClause(order TaskOrder) string {
	switch order {
	case TaskCreatedAtAsc:
		return "created_at ASC"
	case TaskCreatedAtDesc:
		return "created_at DESC"
	case TaskDueDateAsc:
		return "due_date ASC"
	case TaskDueDateDesc:
		return "due_date DESC"
	case TaskSubjectAsc:
		return "subject ASC"
	case TaskSubjectDesc:
		return "subject DESC"
	}

	return "id ASC"
}
