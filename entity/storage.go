package entity

import (
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	AddTask(subject string, dueDate time.Time, description string) Task
	DeleteTask(id uuid.UUID)
	Tasks(order string) []Task
}

type inMemoryStorage struct {
	tasks []Task
	sync.RWMutex
}

func NewInMemory() Storage {
	return &inMemoryStorage{
		tasks: []Task{},
	}
}

func (s *inMemoryStorage) DeleteTask(id uuid.UUID) {
	s.Lock()
	defer s.Unlock()
	s.tasks = slices.DeleteFunc(s.tasks, func(t Task) bool {
		return t.Id == id
	})
}

func (s *inMemoryStorage) AddTask(subject string, dueDate time.Time, description string) Task {
	s.Lock()
	defer s.Unlock()
	task := Task{
		Id:         uuid.New(),
		Subject:    subject,
		CreatedAt:  time.Now(),
		DueDate:    dueDate,
		Desciption: description,
	}
	s.tasks = append(s.tasks, task)
	return task
}

func (s *inMemoryStorage) Tasks(order string) []Task {
	s.RLock()
	defer s.RUnlock()
	t := slices.Clone(s.tasks)
	slices.SortStableFunc(t, TaskOrderFunc(order))
	return t
}
