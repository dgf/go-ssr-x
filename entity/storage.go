package entity

import (
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	AddTask(subject string, dueDate time.Time, description string) uuid.UUID
	DeleteTask(id uuid.UUID)
	Tasks(order string) []Task
}
