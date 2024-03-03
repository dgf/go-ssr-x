package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/locales"
	"github.com/dgf/go-ssr-x/server"
	"github.com/dgf/go-ssr-x/view"
)

//go:embed assets/*
var assets embed.FS

var (
	mux     *http.ServeMux
	storage entity.Storage
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	mux = http.NewServeMux()
	mux.Handle("/assets/", http.FileServer(http.FS(assets)))
	mux.HandleFunc("DELETE /clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
	})

	storage = entity.NewMemory()
	for i := range 100 {
		dueInDays := time.Duration(i%14) * 24 * time.Hour // mods a day in the next two weeks
		subject := fmt.Sprintf("to do %v something", i+1)
		desc := fmt.Sprintf("# first\n## second\nsome `code` check\n```\nmore\ncode\n```\n\nlist:\n\n- %v", strings.Join([]string{"foo", "bar"}, "\n- "))
		storage.AddTask(subject, time.Now().Add(dueInDays), desc)
	}
}

func route(pattern string, handler func(http.ResponseWriter, *http.Request) templ.Component) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		translator := locales.NewTranslator(r)
		ctx := context.WithValue(r.Context(), view.TranslatorKey, translator)

		component := handler(w, r)
		if r.Header.Get("HX-Request") == "true" {
			component.Render(ctx, w)
		} else {
			view.Page(component).Render(ctx, w)
		}
	})
}

func main() {
	taskServer := server.NewTaskServer(storage)

	route("GET /tasks/new", taskServer.TaskCreateForm)
	route("GET /tasks/rows", taskServer.TaskRows)
	route("GET /tasks", taskServer.TasksSection)
	route("POST /tasks", taskServer.CreateTask)
	route("GET /tasks/{id}", taskServer.ShowTask)
	route("GET /tasks/{id}/edit", taskServer.EditTask)
	route("DELETE /tasks/{id}", taskServer.DeleteTask)
	route("PUT /tasks/{id}", taskServer.UpdateTask)

	route("/", func(w http.ResponseWriter, r *http.Request) templ.Component {
		if r.URL.Path == "/" {
			return taskServer.TasksSection(w, r)
		}

		w.WriteHeader(404)
		return view.ClientError("not_found_path", map[string]string{"method": r.Method, "path": r.URL.Path})
	})

	slog.Info("Listening on :3000")
	http.ListenAndServe("0.0.0.0:3000", mux)
}
