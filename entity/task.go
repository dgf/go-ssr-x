package entity

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Id         uuid.UUID
	CreatedAt  time.Time
	DueDate    time.Time
	Subject    string
	Desciption string
}

type TaskOverview struct {
	CreatedAt time.Time
	DueDate   time.Time
	Subject   string
	Id        uuid.UUID
}

type TaskSort int64

const (
	TaskSortCreatedAt TaskSort = iota
	TaskSortDueDate
	TaskSortSubject
	TaskSortDefault = TaskSortDueDate
)

var taskSortKeys = []string{
	"created-at",
	"due-date",
	"subject",
}

func (o TaskSort) String() string {
	return taskSortKeys[o]
}

func TaskSortOrDefault(sort string) TaskSort {
	o := slices.Index(taskSortKeys, sort)
	if o == -1 {
		return TaskSortDefault
	}
	return TaskSort(o)
}
