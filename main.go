package main

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/server"
	"github.com/dgf/go-ssr-x/view"
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
		desc := fmt.Sprintf("# first\n## second\nsome `code` check\n```\nmore\ncode\n```\n\nlist:\n\n- %v", strings.Join([]string{"foo", "bar"}, "\n- "))
		storage.AddTask(subject, time.Now().Add(dueInDays), desc)
	}
}

func main() {
	mux := http.NewServeMux()
	taskServer := server.NewTaskServer(storage, defaultTaskOrder)

	mux.Handle("/assets/", http.FileServer(http.FS(assets)))
	mux.HandleFunc("DELETE /clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
	})

	mux.HandleFunc("GET /tasks/new", taskServer.TaskCreateForm)
	mux.HandleFunc("GET /tasks/rows", taskServer.TaskRows)
	mux.HandleFunc("GET /tasks", taskServer.TasksSection)
	mux.HandleFunc("POST /tasks", taskServer.CreateTask)
	mux.HandleFunc("GET /tasks/{id}", taskServer.Task)
	mux.HandleFunc("DELETE /tasks/{id}", taskServer.DeleteTask)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			view.IndexPage(storage.Tasks(defaultTaskOrder), defaultTaskOrder).Render(r.Context(), w)
			return
		}

		w.WriteHeader(404)
		view.NotFoundPage(r.Method, r.URL.Path).Render(r.Context(), w)
	})

	fmt.Println("Listening on :3000")
	http.ListenAndServe("0.0.0.0:3000", mux)
}
