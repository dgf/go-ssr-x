package storage

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	CreatedAt  time.Time
	DoneAt     time.Time
	DueDate    time.Time
	Subject    string
	Desciption string
	Id         uuid.UUID
}
