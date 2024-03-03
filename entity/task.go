package entity

import (
	"cmp"
	"slices"
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

type TaskOrder int64

const (
	CreatedAtAsc TaskOrder = iota
	CreatedAtDesc
	DueDateAsc
	DueDateDesc
	SubjectAsc
	SubjectDesc
	DefaultTaskOrder = DueDateAsc
)

var taskOrderLabels = []string{
	"created-asc",
	"created-desc",
	"due-date-asc",
	"due-date-desc",
	"subject-asc",
	"subject-desc",
}

func (o TaskOrder) String() string {
	return taskOrderLabels[o]
}

func TaskOrderOrDefault(order string) TaskOrder {
	o := slices.Index(taskOrderLabels, order)
	if o == -1 {
		return DefaultTaskOrder
	}
	return TaskOrder(o)
}

func taskOrderFunc(order TaskOrder) func(i, j Task) int {
	switch order {
	case CreatedAtAsc:
		return func(i, j Task) int {
			return cmp.Compare(i.CreatedAt.String(), j.CreatedAt.String())
		}
	case CreatedAtDesc:
		return func(i, j Task) int {
			return cmp.Compare(j.CreatedAt.String(), i.CreatedAt.String())
		}
	case DueDateAsc:
		return func(i, j Task) int {
			return cmp.Compare(i.DueDate.String(), j.DueDate.String())
		}
	case DueDateDesc:
		return func(i, j Task) int {
			return cmp.Compare(j.DueDate.String(), i.DueDate.String())
		}
	case SubjectAsc:
		return func(i, j Task) int {
			return cmp.Compare(i.Subject, j.Subject)
		}
	case SubjectDesc:
		return func(i, j Task) int {
			return cmp.Compare(j.Subject, i.Subject)
		}
	}

	return func(i, j Task) int {
		return cmp.Compare(i.Id.String(), j.Id.String())
	}
}
