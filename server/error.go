package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/dgf/go-ssr-x/view"
)

func clientError(w http.ResponseWriter, statusCode int, message string) templ.Component {
	w.Header().Add("HX-Reswap", "afterbegin")
	w.WriteHeader(statusCode)
	return view.ClientErrorNotify(statusCode, message)
}
