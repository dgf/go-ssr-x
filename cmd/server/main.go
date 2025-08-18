package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/log"
	"github.com/dgf/go-ssr-x/postgres"
	"github.com/dgf/go-ssr-x/sqlite3"
	"github.com/dgf/go-ssr-x/web"
	"golang.org/x/exp/slices"

	_ "modernc.org/sqlite"
)

const (
	defaultAddr       = "0.0.0.0:3000"
	defaultConnStr    = "postgres://task-db-user:my53cr3tpa55w0rd@localhost?sslmode=disable"
	defaultSqliteFile = ".tasks.sqlite"
)

func parseFlags(ctx context.Context, server *web.Server) error {
	var addr, connStr, storage string

	flag.StringVar(&addr, "address", defaultAddr, "web server address")
	flag.StringVar(&storage, "storage", "memory", "memory, file or database")
	flag.StringVar(&connStr, "connection", defaultConnStr, "database connection string")
	flag.Parse()

	if !slices.Contains([]string{"memory", "file", "database"}, storage) {
		flag.Usage()

		return fmt.Errorf("unknown storage type: %s", storage)
	}

	server.Addr = addr

	switch storage {
	case "database":
		{
			database, err := postgres.NewDatabase(ctx, connStr)
			if err != nil {
				return err
			}

			server.Storage = database
		}
	case "file":
		{
			if connStr == defaultConnStr {
				connStr = defaultSqliteFile
			}
			log.Info("use file storage", "path", connStr)

			file, err := sqlite3.NewFile(ctx, connStr)
			if err != nil {
				return err
			}

			server.Storage = file
		}
	default:
		{
			log.Warn("running with in-memory storage, the data will be lost when restarting")

			server.Storage = entity.NewMemory()
		}
	}

	return nil
}

func initStorage(ctx context.Context, storage entity.Storage) error {
	taskCount, err := storage.TaskCount(ctx)
	if err != nil {
		return fmt.Errorf("initial storage access failed: %w", err)
	}

	if taskCount == 0 {
		log.Info("initialize storage with some tasks")
		for i := range 100 {
			_, err := storage.AddTask(ctx, entity.TaskData{
				DueDate:     time.Now().Add(time.Duration(i%14) * 24 * time.Hour), // mods a day in the next two weeks
				Subject:     fmt.Sprintf("to do %v something", i+1),
				Description: "some `code` check\n\nlist:\n\n- foo\n- bar",
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	ctx := context.Background()

	server := web.NewServer()
	err := parseFlags(ctx, server)
	if err != nil {
		panic(err)
	}

	err = initStorage(ctx, server.Storage)
	if err != nil {
		panic(err)
	}
	defer server.Storage.Close()

	log.Info("Listening on " + server.Addr)
	log.Error("listen and serve failed", server.Serve())
}
