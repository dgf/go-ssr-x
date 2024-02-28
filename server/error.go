package server

import (
	"context"
	"net/http"

	"github.com/dgf/go-ssr-x/view"
)

func clientError(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("HX-Reswap", "afterbegin")
	w.WriteHeader(statusCode)
	view.ClientErrorNotify(statusCode, message).Render(ctx, w)
}
