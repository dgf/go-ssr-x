package entity

import (
	"cmp"
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

func (m *memory) AddTask(dueDate time.Time, subject, description string) (uuid.UUID, error) {
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
	return id, nil
}

func (m *memory) TaskCount() (int, error) {
	m.RLock()
	defer m.RUnlock()

	return len(m.tasks), nil
}

func (m *memory) Task(id uuid.UUID) (Task, bool, error) {
	m.RLock()
	defer m.RUnlock()

	t, ok := m.tasks[id]
	return t, ok, nil
}

func (m *memory) Tasks(order TaskOrder) ([]Task, error) {
	m.RLock()
	defer m.RUnlock()

	p := make([]Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		p = append(p, t)
	}

	slices.SortStableFunc(p, taskOrderFunc(order))
	return p, nil
}

func (m *memory) DeleteTask(id uuid.UUID) error {
	m.Lock()
	defer m.Unlock()

	delete(m.tasks, id)
	return nil
}

func (m *memory) UpdateTask(id uuid.UUID, dueDate time.Time, subject, description string) (Task, bool, error) {
	m.Lock()
	defer m.Unlock()

	if t, ok := m.tasks[id]; !ok {
		return t, false, nil
	} else {
		t.Subject = subject
		t.DueDate = dueDate
		t.Desciption = description
		m.tasks[id] = t
		return t, true, nil
	}
}

func taskOrderFunc(order TaskOrder) func(i, j Task) int {
	switch order {
	case TaskCreatedAtAsc:
		return func(i, j Task) int {
			return cmp.Compare(i.CreatedAt.String(), j.CreatedAt.String())
		}
	case TaskCreatedAtDesc:
		return func(i, j Task) int {
			return cmp.Compare(j.CreatedAt.String(), i.CreatedAt.String())
		}
	case TaskDueDateAsc:
		return func(i, j Task) int {
			return cmp.Compare(i.DueDate.String(), j.DueDate.String())
		}
	case TaskDueDateDesc:
		return func(i, j Task) int {
			return cmp.Compare(j.DueDate.String(), i.DueDate.String())
		}
	case TaskSubjectAsc:
		return func(i, j Task) int {
			return cmp.Compare(i.Subject, j.Subject)
		}
	case TaskSubjectDesc:
		return func(i, j Task) int {
			return cmp.Compare(j.Subject, i.Subject)
		}
	}

	return func(i, j Task) int {
		return cmp.Compare(i.Id.String(), j.Id.String())
	}
}
