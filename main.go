package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/locale"
	"github.com/dgf/go-ssr-x/server"
	"github.com/dgf/go-ssr-x/view"
)

//go:embed assets/*
var assets embed.FS

var (
	mux         *http.ServeMux
	storage     entity.Storage
	storageType string
	connStr     string
)

const defaultConnStr = "postgres://task-db-user:my53cr3tpa55w0rd@localhost?sslmode=disable"

func parseFlags() {
	flag.StringVar(&storageType, "storage", "memory", "memory or database")
	flag.StringVar(&connStr, "connection", defaultConnStr, "database connection string")
	flag.Parse()

	if storageType != "memory" && storageType != "database" {
		flag.Usage()
		os.Exit(1)
	}
}

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	mux = http.NewServeMux()
	mux.Handle("/assets/", http.FileServer(http.FS(assets)))
	mux.HandleFunc("DELETE /clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
	})
}

func route(pattern string, handler func(http.ResponseWriter, *http.Request) templ.Component) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		ctx := context.WithValue(r.Context(), view.LocaleHelperKey, view.LocaleHelper{
			Formatter:  locale.NewFormatter(r),
			Translator: locale.NewTranslator(r),
		})

		component := handler(w, r)
		if r.Header.Get("HX-Request") != "true" {
			component = view.Page(component)
		}
		if err := component.Render(ctx, w); err != nil {
			slog.Error(fmt.Sprintf("component rendering failed: %v", err))
		}
	})
}

func main() {
	parseFlags()
	if storageType == "memory" {
		slog.Warn("running with in-memory storage, the data will be lost when restarting")
		storage = entity.NewMemory()
	}

	if storageType == "database" {
		storage = entity.NewDatabase(connStr)
	}

	if taskCount, err := storage.TaskCount(); err != nil {
		slog.Error(fmt.Sprintf("initial database query failed: %v", err))
		os.Exit(7)
	} else if taskCount == 0 {
		slog.Info("initialize storage with some tasks")
		for i := range 100 {
			dueInDays := time.Duration(i%14) * 24 * time.Hour // mods a day in the next two weeks
			subject := fmt.Sprintf("to do %v something", i+1)
			desc := "some `code` check\n\nlist:\n\n- foo\n- bar"
			_, _ = storage.AddTask(time.Now().Add(dueInDays), subject, desc)
		}
	}

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
	slog.Error(fmt.Sprintf("listen and serve failed: %v", http.ListenAndServe("0.0.0.0:3000", mux)))
}
