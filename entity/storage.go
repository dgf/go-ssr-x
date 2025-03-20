package entity

import (
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	AddTask(dueDate time.Time, subject, description string) (uuid.UUID, error)
	TaskCount() (int, error)
	Task(id uuid.UUID) (task Task, found bool, err error)
	Tasks(page TaskPage) ([]TaskOverview, error)
	DeleteTask(id uuid.UUID) error
	UpdateTask(id uuid.UUID, dueDate time.Time, subject, description string) (task Task, found bool, err error)
}
