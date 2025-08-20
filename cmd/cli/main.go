package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/glamour"
	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/locale"
	"github.com/dgf/go-ssr-x/sqlite3"
	"github.com/google/uuid"
	"golang.org/x/text/language"

	_ "modernc.org/sqlite"
)

type Globals struct {
	File string
}

func runWithStorage(globals *Globals, run func(context.Context, io.Writer, entity.Storage) error) error {
	ctx := context.Background()
	ctx = locale.WithLocale(ctx, language.German)

	storage, err := sqlite3.NewFile(ctx, globals.File)
	if err != nil {
		return err
	}

	return run(ctx, os.Stdout, storage)
}

type PageTasksCmd struct {
	Page   int `default:"1" help:"page number to show."`
	Size   int `default:"10" help:"page size to show."`
	Sort   entity.TaskSort
	Order  entity.SortOrder
	Filter string `help:"Match subject filter."`
}

func (cmd *PageTasksCmd) Run(globals *Globals) error {
	return runWithStorage(globals, func(ctx context.Context, w io.Writer, storage entity.Storage) error {
		q := entity.TaskQuery{
			Page:   cmd.Page,
			Size:   cmd.Size,
			Sort:   cmd.Sort,
			Order:  cmd.Order,
			Filter: cmd.Filter,
		}

		page, err := storage.Tasks(ctx, q)
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "\n %s (%d / %d)   %s: %d, %s: %s %s, %s: %s \n\n",
			locale.Translate(ctx, "page_title"),
			page.Results,
			page.Count,
			locale.Translate(ctx, "page_number"),
			q.Page,
			locale.Translate(ctx, "task_sort"),
			q.Order,
			q.Sort,
			locale.Translate(ctx, "task_subject"),
			q.Filter)

		fmt.Fprintf(w, " # \t %s \t %s \t %s \n\n",
			"UUID                                ",
			locale.Translate(ctx, "task_due_date"),
			locale.Translate(ctx, "task_subject"))

		for t, task := range page.Tasks {
			fmt.Fprintf(w, " %d \t %s \t %s \t %s\n",
				page.Start+t+1,
				task.ID,
				locale.LocalizeDate(ctx, task.DueDate),
				task.Subject)
		}

		return nil
	})
}

type ShowTaskCmd struct {
	UUID uuid.UUID `arg:"" required:"" help:"uuid of task to show"`
}

func (cmd *ShowTaskCmd) Run(globals *Globals) error {
	return runWithStorage(globals, func(ctx context.Context, w io.Writer, storage entity.Storage) error {
		task, found, err := storage.Task(ctx, cmd.UUID)
		if err != nil {
			return err
		}

		if !found {
			return errors.New(locale.TranslateData(ctx, "not_found_task", map[string]string{"id": cmd.UUID.String()}))
		}

		fmt.Fprintf(w, "\n %s:\t%s\n",
			locale.Translate(ctx, "task_subject"), task.Subject)
		fmt.Fprintf(w, " %s:\t%s\n\n",
			locale.Translate(ctx, "task_due_date"), locale.LocalizeDate(ctx, task.DueDate))

		r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
		out, err := r.Render(task.Description)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, out)

		return nil
	})
}

var CLI struct {
	File string `default:".tasks.sqlite" help:"File based storage backend path (your data)."`

	Tasks PageTasksCmd `cmd:"" default:"1" help:"List tasks."`
	Show  ShowTaskCmd  `cmd:"" help:"Show task."`
}

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run(&Globals{File: CLI.File})
	ctx.FatalIfErrorf(err)
}
