package view

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/a-h/templ"
	"github.com/dgf/go-ssr-x/log"
	"github.com/yuin/goldmark"
)

func date(d time.Time) string {
	if !d.IsZero() {
		return d.Format(time.DateOnly)
	}
	return ""
}

func markdown(md string) templ.Component {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		log.Warn(fmt.Sprintf("failed to convert markdown to HTML: %v", err))
		return templ.NopComponent
	}

	return templ.ComponentFunc(func(_ context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, buf.String())
		return
	})
}
