// Package web provides a web server that can serve tasks from a storage.
package web

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/locale"
	"github.com/dgf/go-ssr-x/log"
	"github.com/dgf/go-ssr-x/web/view"
	"golang.org/x/text/language"
)

//go:embed assets/*
var assets embed.FS

type server struct {
	addr    string
	storage entity.Storage
	mux     *http.ServeMux
}

func Serve(addr string, storage entity.Storage) error {
	s := server{
		addr:    addr,
		storage: storage,
		mux:     http.NewServeMux(),
	}

	s.mux.Handle("/assets/", http.FileServer(http.FS(assets)))
	s.mux.HandleFunc("DELETE /clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK) // 204 is currently ignored, see https://github.com/bigskysoftware/htmx/issues/2194
	})

	return s.serve()
}

func acceptLanguageOrDefault(r *http.Request) language.Tag {
	accept := r.Header.Get("Accept-Language")
	tags, _, err := language.ParseAcceptLanguage(accept)
	if err != nil {
		log.Info("accept language header parse failed", "header", accept)

		return language.English
	}

	if len(tags) == 0 {
		return language.English
	}

	return tags[0]
}

func (s *server) route(pattern string, handler func(http.ResponseWriter, *http.Request) templ.Component) {
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		lang := acceptLanguageOrDefault(r)
		ctx := context.WithValue(r.Context(), view.LocaleContextKey, view.LocaleContext{
			Formatter:  locale.RequestFormatter(lang),
			Translator: locale.RequestTranslator(lang),
		})

		component := handler(w, r)
		if r.Header.Get("HX-Request") != "true" {
			component = view.Page(component)
		}

		err := component.Render(ctx, w)
		if err != nil {
			log.Error("component rendering failed", err)
		}
	})
}

func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				if e, ok := err.(error); ok {
					log.Error("panic recovery uncaught error", e)
				} else {
					log.Error("panic recovery uncaught error", fmt.Errorf("%v", err))
				}
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func (s *server) serve() error {
	taskServer := NewTaskServer(s.storage)

	s.route("GET /tasks/new", taskServer.TaskCreateForm)
	s.route("GET /tasks/rows", taskServer.TaskRows)
	s.route("GET /tasks", taskServer.TasksSection)
	s.route("POST /tasks", taskServer.CreateTask)
	s.route("GET /tasks/{id}", taskServer.ShowTask)
	s.route("GET /tasks/{id}/edit", taskServer.EditTask)
	s.route("DELETE /tasks/{id}", taskServer.DeleteTask)
	s.route("PUT /tasks/{id}", taskServer.UpdateTask)

	s.route("/", func(w http.ResponseWriter, r *http.Request) templ.Component {
		if r.URL.Path == "/" {
			return taskServer.TasksSection(w, r)
		}

		if r.URL.Path == "/error" {
			panic("don't panic!")
		}

		w.WriteHeader(http.StatusNotFound)

		return view.ClientError("not_found_path", map[string]string{"method": r.Method, "path": r.URL.Path})
	})

	return (&http.Server{
		Addr:         s.addr,
		Handler:      panicRecovery(s.mux),
		WriteTimeout: 13 * time.Second,
		ReadTimeout:  17 * time.Second,
		IdleTimeout:  37 * time.Second,
	}).ListenAndServe()
}
