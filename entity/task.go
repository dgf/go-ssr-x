package entity

import (
	"cmp"
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

func TaskOrderFunc(order string) func(i, j Task) int {
	switch order {
	case "created-asc":
		return func(i, j Task) int {
			return cmp.Compare(i.CreatedAt.String(), j.CreatedAt.String())
		}
	case "created-desc":
		return func(i, j Task) int {
			return cmp.Compare(j.CreatedAt.String(), i.CreatedAt.String())
		}
	case "due-date-asc":
		return func(i, j Task) int {
			return cmp.Compare(i.DueDate.String(), j.DueDate.String())
		}
	case "due-date-desc":
		return func(i, j Task) int {
			return cmp.Compare(j.DueDate.String(), i.DueDate.String())
		}
	case "subject-asc":
		return func(i, j Task) int {
			return cmp.Compare(i.Subject, j.Subject)
		}
	case "subject-desc":
		return func(i, j Task) int {
			return cmp.Compare(j.Subject, i.Subject)
		}
	}

	return func(i, j Task) int {
		return cmp.Compare(i.Id.String(), j.Id.String())
	}
}
