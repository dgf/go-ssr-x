package entity

import (
	"context"

	"github.com/google/uuid"
)

type Storage interface {
	AddTask(ctx context.Context, data TaskData) (uuid.UUID, error)
	TaskCount(ctx context.Context) (int, error)
	Task(ctx context.Context, id uuid.UUID) (task Task, found bool, err error)
	Tasks(ctx context.Context, page TaskQuery) (TaskPage, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	UpdateTask(ctx context.Context, id uuid.UUID, data TaskData) (task Task, found bool, err error)
}
