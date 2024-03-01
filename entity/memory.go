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

func (m *memory) Task(id uuid.UUID) (Task, bool) {
	m.RLock()
	defer m.RUnlock()

	t, ok := m.tasks[id]
	return t, ok
}

func (m *memory) Tasks(order TaskOrder) []Task {
	m.RLock()
	defer m.RUnlock()

	p := make([]Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		p = append(p, t)
	}

	slices.SortStableFunc(p, taskOrderFunc(order))
	return p
}

func (m *memory) DeleteTask(id uuid.UUID) {
	m.Lock()
	defer m.Unlock()

	delete(m.tasks, id)
}

func (m *memory) UpdateTask(id uuid.UUID, subject string, dueDate time.Time, description string) (Task, bool) {
	m.Lock()
	defer m.Unlock()

	if t, ok := m.tasks[id]; !ok {
		return t, ok
	} else {
		t.Subject = subject
		t.DueDate = dueDate
		t.Desciption = description
		m.tasks[id] = t
		return t, ok
	}
}
