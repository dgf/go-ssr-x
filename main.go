package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/view"
	"github.com/google/uuid"
)

//go:embed assets/*
var assets embed.FS
var storage entity.Storage

const defaultTaskOrder = "due-date-asc"

func init() {
	storage = entity.NewMemory()

	for i := range 100 {
		dueInDays := time.Duration(i%14) * 24 * time.Hour // mods a day in the next two weeks
		subject := fmt.Sprintf("to do %v something", i+1)
		desc := fmt.Sprintf("list:\n\n- %v", strings.Join([]string{"foo", "bar"}, "\n- "))
		storage.AddTask(subject, time.Now().Add(dueInDays), desc)
	}
}

func parseDate(date string) time.Time {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return time.Time{}
	}
	return t
}

func clientError(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("HX-Reswap", "afterbegin")
	w.WriteHeader(statusCode)
	view.ClientErrorNotify(statusCode, message).Render(ctx, w)
}

func main() {
	http.Handle("/assets/", http.FileServer(http.FS(assets)))

	http.HandleFunc("DELETE /clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
	})

	http.HandleFunc("GET /tasks/new", func(w http.ResponseWriter, r *http.Request) {
		view.TaskCreateForm().Render(r.Context(), w)
	})

	http.HandleFunc("GET /tasks/rows", func(w http.ResponseWriter, r *http.Request) {
		view.TaskRows(storage.Tasks(r.URL.Query().Get("order"))).Render(r.Context(), w)
	})

	http.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		view.TasksSection(storage.Tasks(defaultTaskOrder), defaultTaskOrder).Render(r.Context(), w)
	})

	http.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		dueDate := parseDate(r.FormValue("dueDate"))
		subject := r.FormValue("subject")
		description := r.FormValue("description")

		id := storage.AddTask(subject, dueDate, description)
		message := fmt.Sprintf("task %q created", id)
		view.TasksSectionWithNotifyOOB(storage.Tasks(defaultTaskOrder), defaultTaskOrder, message).Render(r.Context(), w)
	})

	http.HandleFunc("DELETE /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		pid := r.PathValue("id")
		if id, err := uuid.Parse(pid); err != nil {
			clientError(r.Context(), w, http.StatusBadRequest, fmt.Sprintf("invalid param %q", pid))
		} else if !storage.HasTask(id) {
			clientError(r.Context(), w, http.StatusNotFound, fmt.Sprintf("task %q", id))
		} else {
			storage.DeleteTask(id)
			w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			view.IndexPage(storage.Tasks(defaultTaskOrder), defaultTaskOrder).Render(r.Context(), w)
			return
		}

		w.WriteHeader(404)
		view.NotFoundPage(r.Method, r.URL.Path).Render(r.Context(), w)
	})

	fmt.Println("Listening on :3000")
	http.ListenAndServe("0.0.0.0:3000", nil)
}
