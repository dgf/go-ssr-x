package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/log"
	"github.com/dgf/go-ssr-x/web"
)

const defaultConnStr = "postgres://task-db-user:my53cr3tpa55w0rd@localhost?sslmode=disable"

var (
	storage     entity.Storage
	storageType string
	connStr     string
)

func parseFlags() {
	flag.StringVar(&storageType, "storage", "memory", "memory or database")
	flag.StringVar(&connStr, "connection", defaultConnStr, "database connection string")
	flag.Parse()

	if storageType != "memory" && storageType != "database" {
		flag.Usage()
		os.Exit(1)
	}
}

func initStorage(ctx context.Context) entity.Storage {
	switch storageType {
	case "memory":
		log.Warn("running with in-memory storage, the data will be lost when restarting")
		return entity.NewMemory()
	case "database":
		if storage, err := entity.NewDatabase(ctx, connStr); err != nil {
			panic(err)
		} else {
			return storage
		}
	default:
		panic(fmt.Sprintf("unknown storage type: %s", storageType))
	}
}

func main() {
	ctx := context.Background()
	parseFlags()

	storage := initStorage(ctx)
	defer storage.Close()

	if taskCount, err := storage.TaskCount(ctx); err != nil {
		log.Error("initial storage access failed", err)
		os.Exit(7)
	} else if taskCount == 0 {
		log.Info("initialize storage with some tasks")
		for i := range 100 {
			_, _ = storage.AddTask(ctx, entity.TaskData{
				DueDate:     time.Now().Add(time.Duration(i%14) * 24 * time.Hour), // mods a day in the next two weeks
				Subject:     fmt.Sprintf("to do %v something", i+1),
				Description: "some `code` check\n\nlist:\n\n- foo\n- bar",
			})
		}
	}

	addr := "0.0.0.0:3000"

	log.Info("Listening on " + addr)
	log.Error("listen and serve failed", web.Serve(addr, storage))
}
