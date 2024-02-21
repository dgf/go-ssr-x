package main

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/dgf/go-ssr-x/storage"
	"github.com/dgf/go-ssr-x/view"
)

//go:embed assets/*
var assets embed.FS

func parseFormDate(date string) time.Time {
	fmt.Println("check date", date)
	if len(date) == 0 {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}
	}
	return t
}

func main() {
	s := storage.New()

	http.Handle("/assets/", http.FileServer(http.FS(assets)))

	http.HandleFunc("GET /tasks/new", func(w http.ResponseWriter, r *http.Request) {
		view.TaskCreateForm().Render(r.Context(), w)
	})

	http.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		view.TasksSection(s.Tasks()).Render(r.Context(), w)
	})

	http.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		// time.Sleep(1 * time.Second)
		dueDate := r.FormValue("dueDate")
		s.AddTask(r.FormValue("subject"), parseFormDate(dueDate), r.FormValue("description"))
		view.TasksSection(s.Tasks()).Render(r.Context(), w)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			view.Index(s.Tasks()).Render(r.Context(), w)
			return
		}

		w.WriteHeader(404)
		view.NotFound(r.Method, r.URL.Path).Render(r.Context(), w)
	})

	fmt.Println("Listening on :3000")
	http.ListenAndServe("0.0.0.0:3000", nil)
}
