package server

import (
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

func (ts *TaskServer) handleTask(w http.ResponseWriter, r *http.Request, handler func(task entity.Task) templ.Component) templ.Component {
	pid := r.PathValue("id")
	if id, err := uuid.Parse(pid); err != nil {
		return clientError(w, http.StatusBadRequest, "bad_request_path_param", map[string]string{"param": "id", "value": pid})
	} else if task, ok := ts.storage.Task(id); !ok {
		return clientError(w, http.StatusNotFound, "not_found_task", map[string]string{"id": pid})
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
	w.Header().Add("HX-Push-Url", "/tasks?order="+order.String())
	return view.TasksSection(ts.storage.Tasks(order), order)
}

func (ts *TaskServer) CreateTask(w http.ResponseWriter, r *http.Request) templ.Component {
	r.ParseForm()
	dueDate := parseDate(r.FormValue("dueDate"))
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	id := ts.storage.AddTask(subject, dueDate, description)
	return view.TasksSectionWithNotifyOOB(ts.storage.Tasks(entity.DefaultTaskOrder), entity.DefaultTaskOrder, "ok_task_created", map[string]string{"id": id.String()})
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
	r.ParseForm()
	dueDate := parseDate(r.FormValue("dueDate"))
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	return ts.handleTask(w, r, func(task entity.Task) templ.Component {
		if updated, ok := ts.storage.UpdateTask(task.Id, subject, dueDate, description); !ok {
			return clientError(w, http.StatusConflict, "conflict_task_update", nil)
		} else {
			return view.TaskDetailsWithNotifyOOB(updated, "ok_task_updated", map[string]string{"id": string(updated.Id.String())})
		}
	})
}
