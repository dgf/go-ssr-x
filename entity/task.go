package entity

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type TaskOverview struct {
	CreatedAt time.Time
	DueDate   time.Time
	Subject   string
	Id        uuid.UUID
}

type Task struct {
	DueDate    time.Time
	Subject    string
	CreatedAt  time.Time
	Desciption string
	Id         uuid.UUID
}

type TaskOrder int64

const (
	TaskCreatedAtAsc TaskOrder = iota
	TaskCreatedAtDesc
	TaskDueDateAsc
	TaskDueDateDesc
	TaskSubjectAsc
	TaskSubjectDesc
	TaskDefaultOrder = TaskDueDateAsc
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
		return TaskDefaultOrder
	}
	return TaskOrder(o)
}
