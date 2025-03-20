package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/a-h/templ"
	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/log"
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
	} else if task, ok, err := ts.storage.Task(id); err != nil {
		log.Error("task access failed", err)
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	} else if !ok {
		idData := map[string]string{"id": pid}
		return clientError(w, r, http.StatusNotFound, "not_found_task", idData)
	} else {
		return handler(task)
	}
}

func query2TaskPage(query url.Values) entity.TaskPage {
	return entity.TaskPage{
		Page:   param2IntOrDefault(query, "page", 1),
		Size:   param2IntOrDefault(query, "size", entity.TaskPageDefaultSize),
		Sort:   entity.TaskSortOrDefault(query.Get("sort")),
		Order:  entity.SortOrderOrDefault(query.Get("order")),
		Filter: query.Get("subject"),
	}
}

func taskPage2Query(page entity.TaskPage) string {
	query := &url.Values{}

	query.Add("sort", page.Sort.String())
	query.Add("order", page.Order.String())
	query.Add("subject", page.Filter)
	query.Add("page", strconv.Itoa(page.Page))
	query.Add("size", strconv.Itoa(page.Size))

	return query.Encode()
}

func (ts *TaskServer) TaskCreateForm(w http.ResponseWriter, r *http.Request) templ.Component {
	return view.TaskCreateForm()
}

func (ts *TaskServer) TaskRows(w http.ResponseWriter, r *http.Request) templ.Component {
	page := query2TaskPage(r.URL.Query())

	pushURL := &url.URL{
		Path:     "/tasks",
		RawQuery: taskPage2Query(page),
	}

	w.Header().Add("HX-Push-Url", pushURL.String())

	if tasks, err := ts.storage.Tasks(page); err != nil {
		log.Error("task rows access failed", err)
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	} else {
		return view.TaskRows(tasks)
	}
}

func (ts *TaskServer) TasksSection(w http.ResponseWriter, r *http.Request) templ.Component {
	page := query2TaskPage(r.URL.Query())

	if tasks, err := ts.storage.Tasks(page); err != nil {
		log.Error("tasks section access failed", err)
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	} else {
		return view.TasksSection(tasks, page)
	}
}

func (ts *TaskServer) CreateTask(w http.ResponseWriter, r *http.Request) templ.Component {
	if err := r.ParseForm(); err != nil {
		log.Warn(fmt.Sprintf("task create form parsing failed: %v", err))
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	}

	dueDate := parseDate(r.FormValue("dueDate"))
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	id, err := ts.storage.AddTask(dueDate, subject, description)
	if err != nil {
		log.Error("task creation failed", err)
		messageData := map[string]string{"message": err.Error()}
		return clientError(w, r, http.StatusInternalServerError, "database_error", messageData)
	}

	page := entity.TaskPage{
		Page:   1,
		Size:   entity.TaskPageDefaultSize,
		Sort:   entity.TaskSortDefault,
		Order:  entity.AscendingOrder,
		Filter: "",
	}

	if tasks, err := ts.storage.Tasks(page); err != nil {
		log.Error("task listing failed", err)
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	} else {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			idData := map[string]string{"id": id.String()}
			if err := view.SuccessNotify("ok_task_created", idData).Render(ctx, w); err != nil {
				return err
			}
			return view.TasksSection(tasks, page).Render(ctx, w)
		})
	}
}

func (ts *TaskServer) ShowTask(w http.ResponseWriter, r *http.Request) templ.Component {
	return ts.handleTask(w, r, view.TaskDetails)
}

func (ts *TaskServer) EditTask(w http.ResponseWriter, r *http.Request) templ.Component {
	return ts.handleTask(w, r, view.TaskEditForm)
}

func (ts *TaskServer) DeleteTask(w http.ResponseWriter, r *http.Request) templ.Component {
	return ts.handleTask(w, r, func(task entity.Task) templ.Component {
		if err := ts.storage.DeleteTask(task.Id); err != nil {
			log.Warn(fmt.Sprintf("task deletion failed: %v", err))
			return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
		}
		w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
		return templ.NopComponent
	})
}

func (ts *TaskServer) UpdateTask(w http.ResponseWriter, r *http.Request) templ.Component {
	if err := r.ParseForm(); err != nil {
		log.Warn(fmt.Sprintf("task update form parsing failed: %v", err))
		return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
	}

	dueDate := parseDate(r.FormValue("dueDate"))
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	return ts.handleTask(w, r, func(task entity.Task) templ.Component {
		if updated, ok, err := ts.storage.UpdateTask(task.Id, dueDate, subject, description); err != nil {
			log.Warn(fmt.Sprintf("task update failed: %v ", err))
			return clientError(w, r, http.StatusInternalServerError, "internal_server_error", nil)
		} else if !ok { // e.g. delete while user is updating
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
