package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/view"
	"github.com/google/uuid"
)

type TaskServer struct {
	storage entity.Storage
}

func NewTaskServer(storage entity.Storage) *TaskServer {
	return &TaskServer{storage: storage}
}

type handlerFunc func(entity.Task) templ.Component

func (ts *TaskServer) handleTask(w http.ResponseWriter, r *http.Request, handler handlerFunc) templ.Component {
	pid := r.PathValue("id")
	if id, err := uuid.Parse(pid); err != nil {
		badData := map[string]string{"param": "id", "value": pid}
		return clientError(w, r, http.StatusBadRequest, "bad_request_path_param", badData)
	} else if task, ok := ts.storage.Task(id); !ok {
		idData := map[string]string{"id": pid}
		return clientError(w, r, http.StatusNotFound, "not_found_task", idData)
	} else {
		return handler(task)
	}
}

func (ts *TaskServer) TaskCreateForm(w http.ResponseWriter, r *http.Request) templ.Component {
	return view.TaskCreateForm()
}

func (ts *TaskServer) TaskRows(w http.ResponseWriter, r *http.Request) templ.Component {
	order := entity.TaskOrderOrDefault(r.URL.Query().Get("order"))
	w.Header().Add("HX-Push-Url", "/tasks?order="+order.String())
	return view.TaskRows(ts.storage.Tasks(order))
}

func (ts *TaskServer) TasksSection(w http.ResponseWriter, r *http.Request) templ.Component {
	order := entity.TaskOrderOrDefault(r.URL.Query().Get("order"))
	return view.TasksSection(ts.storage.Tasks(order), order)
}

func (ts *TaskServer) CreateTask(w http.ResponseWriter, r *http.Request) templ.Component {
	if err := r.ParseForm(); err != nil {
		slog.Warn(fmt.Sprintf("task create form parsing failed: %v ", err))
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	}

	dueDate := parseDate(r.FormValue("dueDate"))
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	id := ts.storage.AddTask(subject, dueDate, description)
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		idData := map[string]string{"id": id.String()}
		if err := view.SuccessNotify("ok_task_created", idData).Render(ctx, w); err != nil {
			return err
		}
		return view.TasksSection(ts.storage.Tasks(entity.DefaultTaskOrder), entity.DefaultTaskOrder).Render(ctx, w)
	})
}

func (ts *TaskServer) ShowTask(w http.ResponseWriter, r *http.Request) templ.Component {
	return ts.handleTask(w, r, func(task entity.Task) templ.Component {
		return view.TaskDetails(task)
	})
}

func (ts *TaskServer) EditTask(w http.ResponseWriter, r *http.Request) templ.Component {
	return ts.handleTask(w, r, func(task entity.Task) templ.Component {
		return view.TaskEditForm(task)
	})
}

func (ts *TaskServer) DeleteTask(w http.ResponseWriter, r *http.Request) templ.Component {
	return ts.handleTask(w, r, func(task entity.Task) templ.Component {
		ts.storage.DeleteTask(task.Id)
		w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
		return templ.NopComponent
	})
}

func (ts *TaskServer) UpdateTask(w http.ResponseWriter, r *http.Request) templ.Component {
	if err := r.ParseForm(); err != nil {
		slog.Warn(fmt.Sprintf("task update form parsing failed: %v ", err))
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	}

	dueDate := parseDate(r.FormValue("dueDate"))
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	return ts.handleTask(w, r, func(task entity.Task) templ.Component {
		if updated, ok := ts.storage.UpdateTask(task.Id, subject, dueDate, description); !ok {
			return clientError(w, r, http.StatusConflict, "conflict_task_update", nil)
		} else {
			return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				idData := map[string]string{"id": updated.Id.String()}
				if err := view.SuccessNotify("ok_task_updated", idData).Render(ctx, w); err != nil {
					return err
				}
				return view.TaskDetails(updated).Render(ctx, w)
			})
		}
	})
}
