package storage

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/google/uuid"
)

type memory struct {
	tasks map[uuid.UUID]entity.Task
	sync.RWMutex
}

func NewMemory() entity.Storage {
	return &memory{
		tasks: map[uuid.UUID]entity.Task{},
	}
}

func (m *memory) Close() error {
	return nil
}

func (m *memory) AddTask(ctx context.Context, data entity.TaskData) (uuid.UUID, error) {
	m.Lock()
	defer m.Unlock()

	id := uuid.New()
	m.tasks[id] = entity.Task{
		Id:          id,
		Subject:     data.Subject,
		CreatedAt:   time.Now(),
		DueDate:     data.DueDate,
		Description: data.Description,
	}
	return id, nil
}

func (m *memory) TaskCount(ctx context.Context) (int, error) {
	m.RLock()
	defer m.RUnlock()

	return len(m.tasks), nil
}

func (m *memory) Task(ctx context.Context, id uuid.UUID) (entity.Task, bool, error) {
	m.RLock()
	defer m.RUnlock()

	t, ok := m.tasks[id]
	return t, ok, nil
}

func (m *memory) Tasks(ctx context.Context, query entity.TaskQuery) (entity.TaskPage, error) {
	m.RLock()
	defer m.RUnlock()

	tasks := []entity.Task{}
	for _, t := range m.tasks {
		if strings.Contains(t.Subject, query.Filter) {
			tasks = append(tasks, t)
		}
	}

	slices.SortStableFunc(tasks, taskSortFunc(query.Sort, query.Order))
	pageStart := (query.Page - 1) * query.Size
	if pageStart > len(tasks) {
		return entity.TaskPage{}, nil
	}

	page := entity.TaskPage{
		Count:   len(m.tasks),
		Results: len(tasks),
		Start:   pageStart,
		Tasks:   []entity.TaskOverview{},
	}

	pageEnd := min(pageStart+query.Size, len(tasks))
	for _, t := range tasks[pageStart:pageEnd] {
		page.Tasks = append(page.Tasks, entity.TaskOverview{
			Id:        t.Id,
			CreatedAt: t.CreatedAt,
			DueDate:   t.DueDate,
			Subject:   t.Subject,
		})
	}

	return page, nil
}

func (m *memory) DeleteTask(ctx context.Context, id uuid.UUID) error {
	m.Lock()
	defer m.Unlock()

	delete(m.tasks, id)
	return nil
}

func (m *memory) UpdateTask(ctx context.Context, id uuid.UUID, data entity.TaskData) (entity.Task, bool, error) {
	m.Lock()
	defer m.Unlock()

	if t, ok := m.tasks[id]; !ok {
		return t, false, nil
	} else {
		t.Subject = data.Subject
		t.DueDate = data.DueDate
		t.Description = data.Description
		m.tasks[id] = t
		return t, true, nil
	}
}

func taskSortValue(sort entity.TaskSort) func(entity.Task) string {
	switch sort {
	case entity.TaskSortCreatedAt:
		return func(t entity.Task) string {
			return t.CreatedAt.String()
		}
	case entity.TaskSortDueDate:
		return func(t entity.Task) string {
			return t.DueDate.String()
		}
	case entity.TaskSortSubject:
		return func(t entity.Task) string {
			return t.Subject
		}
	}

	return func(t entity.Task) string {
		return t.Id.String()
	}
}

func taskSortFunc(sort entity.TaskSort, order entity.SortOrder) func(i, j entity.Task) int {
	value := taskSortValue(sort)
	if order == entity.AscendingOrder {
		return func(i, j entity.Task) int {
			return cmp.Compare(value(i), value(j))
		}
	}
	return func(i, j entity.Task) int {
		return cmp.Compare(value(j), value(i))
	}
}
