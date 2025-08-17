package entity

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	DueDate     time.Time
	Subject     string
	Description string
}

type TaskData struct {
	DueDate     time.Time
	Subject     string
	Description string
}

type TaskOverview struct {
	CreatedAt time.Time
	DueDate   time.Time
	Subject   string
	ID        uuid.UUID
}

type TaskSort int64

type TaskQuery struct {
	Page   int
	Size   int
	Sort   TaskSort
	Order  SortOrder
	Filter string
}

type TaskPage struct {
	Count   int
	Results int
	Start   int
	Tasks   []TaskOverview
}

const (
	TaskSortCreatedAt TaskSort = iota
	TaskSortDueDate
	TaskSortSubject
	TaskSortDefault     = TaskSortDueDate
	TaskPageDefaultSize = 10
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
