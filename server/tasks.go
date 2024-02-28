package server

import (
	"fmt"
	"net/http"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/view"
	"github.com/google/uuid"
)

type TaskServer struct {
	storage      entity.Storage
	defaultOrder string
}

func NewTaskServer(storage entity.Storage, defaultOrder string) *TaskServer {
	return &TaskServer{storage: storage, defaultOrder: defaultOrder}
}

func (ts *TaskServer) TaskCreateForm(w http.ResponseWriter, r *http.Request) {
	view.TaskCreateForm().Render(r.Context(), w)
}

func (ts *TaskServer) TaskRows(w http.ResponseWriter, r *http.Request) {
	view.TaskRows(ts.storage.Tasks(r.URL.Query().Get("order"))).Render(r.Context(), w)
}

func (ts *TaskServer) TasksSection(w http.ResponseWriter, r *http.Request) {
	view.TasksSection(ts.storage.Tasks(ts.defaultOrder), ts.defaultOrder).Render(r.Context(), w)
}

func (ts *TaskServer) CreateTask(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	dueDate := parseDate(r.FormValue("dueDate"))
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	id := ts.storage.AddTask(subject, dueDate, description)
	message := fmt.Sprintf("task %q created", id)
	view.TasksSectionWithNotifyOOB(ts.storage.Tasks(ts.defaultOrder), ts.defaultOrder, message).Render(r.Context(), w)
}

func (ts *TaskServer) DeleteTask(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("id")
	if id, err := uuid.Parse(pid); err != nil {
		clientError(r.Context(), w, http.StatusBadRequest, fmt.Sprintf("invalid param %q", pid))
	} else if !ts.storage.HasTask(id) {
		clientError(r.Context(), w, http.StatusNotFound, fmt.Sprintf("task %q", id))
	} else {
		ts.storage.DeleteTask(id)
		w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
	}
}
