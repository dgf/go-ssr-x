package storage

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	AddTask(subject string, dueDate time.Time, description string) Task
	Tasks() []Task
}

type storage struct {
	tasks []Task
	sync.RWMutex
}

func New() Storage {
	return &storage{
		tasks: []Task{},
	}
}

func (s *storage) AddTask(subject string, dueDate time.Time, description string) Task {
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

func (s *storage) Tasks() []Task {
	s.RLock()
	defer s.RUnlock()
	return s.tasks
}
