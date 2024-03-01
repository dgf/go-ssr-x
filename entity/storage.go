package entity

import (
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	AddTask(subject string, dueDate time.Time, description string) uuid.UUID
	Task(id uuid.UUID) (task Task, ok bool)
	Tasks(order TaskOrder) []Task
	DeleteTask(id uuid.UUID)
	UpdateTask(id uuid.UUID, subject string, dueDate time.Time, description string) (task Task, ok bool)
}
