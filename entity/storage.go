package entity

import (
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	AddTask(subject string, dueDate time.Time, description string) uuid.UUID
	DeleteTask(id uuid.UUID)
	HasTask(id uuid.UUID) bool
	Task(id uuid.UUID) (task Task, ok bool)
	Tasks(order string) []Task
}
