package entity

import (
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type memory struct {
	tasks map[uuid.UUID]Task
	sync.RWMutex
}

func NewMemory() Storage {
	return &memory{
		tasks: map[uuid.UUID]Task{},
	}
}

func (m *memory) AddTask(subject string, dueDate time.Time, description string) uuid.UUID {
	m.Lock()
	defer m.Unlock()

	id := uuid.New()
	m.tasks[id] = Task{
		Id:         id,
		Subject:    subject,
		CreatedAt:  time.Now(),
		DueDate:    dueDate,
		Desciption: description,
	}
	return id
}

func (m *memory) DeleteTask(id uuid.UUID) {
	m.Lock()
	defer m.Unlock()

	delete(m.tasks, id)
}

func (m *memory) HasTask(id uuid.UUID) bool {
	_, ok := m.tasks[id]
	return ok
}

func (m *memory) Task(id uuid.UUID) (Task, bool) {
	t, ok := m.tasks[id]
	return t, ok
}

func (m *memory) Tasks(order string) []Task {
	m.RLock()
	defer m.RUnlock()

	p := make([]Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		p = append(p, t)
	}

	slices.SortStableFunc(p, TaskOrderFunc(order))
	return p
}
