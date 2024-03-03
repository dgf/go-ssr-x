package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/dgf/go-ssr-x/view"
)

func clientError(w http.ResponseWriter, r *http.Request, statusCode int, messageID string, data map[string]string) templ.Component {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Add("HX-Reswap", "afterbegin")
		w.WriteHeader(statusCode)
		return view.ClientErrorNotify(messageID, data)
	} else {
		w.WriteHeader(statusCode)
		return view.ClientError(messageID, data)
	}
}
