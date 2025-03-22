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

func (m *memory) Tasks(query TaskQuery) (TaskPage, error) {
	m.RLock()
	defer m.RUnlock()

	tasks := []Task{}
	for _, t := range m.tasks {
		if strings.Contains(t.Subject, query.Filter) {
			tasks = append(tasks, t)
		}
	}

	slices.SortStableFunc(tasks, taskSortFunc(query.Sort, query.Order))
	pageStart := (query.Page - 1) * query.Size
	if pageStart > len(tasks) {
		return TaskPage{}, nil
	}

	page := TaskPage{
		Count:   len(m.tasks),
		Results: len(tasks),
		Start:   pageStart,
		Tasks:   []TaskOverview{},
	}

	pageEnd := min(pageStart+query.Size, len(tasks))
	for _, t := range tasks[pageStart:pageEnd] {
		page.Tasks = append(page.Tasks, TaskOverview{
			Id:        t.Id,
			CreatedAt: t.CreatedAt,
			DueDate:   t.DueDate,
			Subject:   t.Subject,
		})
	}

	return page, nil
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

func taskSortValue(sort TaskSort) func(Task) string {
	switch sort {
	case TaskSortCreatedAt:
		return func(t Task) string {
			return t.CreatedAt.String()
		}
	case TaskSortDueDate:
		return func(t Task) string {
			return t.DueDate.String()
		}
	case TaskSortSubject:
		return func(t Task) string {
			return t.Subject
		}
	}

	return func(t Task) string {
		return t.Id.String()
	}
}

func taskSortFunc(sort TaskSort, order SortOrder) func(i, j Task) int {
	value := taskSortValue(sort)
	if order == AscendingOrder {
		return func(i, j Task) int {
			return cmp.Compare(value(i), value(j))
		}
	}
	return func(i, j Task) int {
		return cmp.Compare(value(j), value(i))
	}
}
