package entity

import (
	"cmp"
	"slices"
	"strings"
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

func (m *memory) Tasks(page TaskPage) ([]TaskOverview, error) {
	m.RLock()
	defer m.RUnlock()

	tasks := make([]TaskOverview, 0, len(m.tasks))
	for _, t := range m.tasks {
		if strings.Contains(t.Subject, page.Filter) {
			tasks = append(tasks, TaskOverview{
				Id:        t.Id,
				CreatedAt: t.CreatedAt,
				DueDate:   t.DueDate,
				Subject:   t.Subject,
			})
		}
	}

	slices.SortStableFunc(tasks, taskSortFunc(page.Sort, page.Order))

	start := (page.Page - 1) * page.Size
	if start > len(tasks) {
		return []TaskOverview{}, nil
	}

	return tasks[start:min(start+page.Size, len(tasks))], nil
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

func taskSortValue(sort TaskSort) func(TaskOverview) string {
	switch sort {
	case TaskSortCreatedAt:
		return func(t TaskOverview) string {
			return t.CreatedAt.String()
		}
	case TaskSortDueDate:
		return func(t TaskOverview) string {
			return t.DueDate.String()
		}
	case TaskSortSubject:
		return func(t TaskOverview) string {
			return t.Subject
		}
	}

	return func(t TaskOverview) string {
		return t.Id.String()
	}
}

func taskSortFunc(sort TaskSort, order SortOrder) func(i, j TaskOverview) int {
	value := taskSortValue(sort)
	if order == AscendingOrder {
		return func(i, j TaskOverview) int {
			return cmp.Compare(value(i), value(j))
		}
	}
	return func(i, j TaskOverview) int {
		return cmp.Compare(value(j), value(i))
	}
}
