package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/dgf/go-ssr-x/entity"
	"github.com/dgf/go-ssr-x/log"
	"github.com/dgf/go-ssr-x/web"
)

type StorageType int

type StorageConfig struct {
	Type    StorageType
	ConnStr string
}

type ServerConfig struct {
	Addr    string
	Storage StorageConfig
}

const (
	MemoryStorage StorageType = iota
	DatabaseStorage

	defaultAddr    = "0.0.0.0:3000"
	defaultConnStr = "postgres://task-db-user:my53cr3tpa55w0rd@localhost?sslmode=disable"
)

func parseFlags() (ServerConfig, error) {
	var addr, connStr, storage string

	flag.StringVar(&addr, "address", defaultAddr, "web server address")
	flag.StringVar(&storage, "storage", "memory", "memory or database")
	flag.StringVar(&connStr, "connection", defaultConnStr, "database connection string")
	flag.Parse()

	config := ServerConfig{Addr: addr, Storage: StorageConfig{Type: MemoryStorage, ConnStr: connStr}}

	if storage != "memory" && storage != "database" {
		flag.Usage()
		return config, fmt.Errorf("unknown storage type: %s", storage)
	} else if storage == "database" {
		config.Storage.Type = DatabaseStorage
	}

	return config, nil
}

func createStorage(ctx context.Context, config ServerConfig) (entity.Storage, error) {
	switch config.Storage.Type {
	case MemoryStorage:
		log.Warn("running with in-memory storage, the data will be lost when restarting")
		return entity.NewMemory(), nil
	case DatabaseStorage:
		return entity.NewDatabase(ctx, config.Storage.ConnStr)
	default:
		return nil, fmt.Errorf("unknown storage type: %d", config.Storage.Type)
	}
}

func initStorage(ctx context.Context, storage entity.Storage) error {
	if taskCount, err := storage.TaskCount(ctx); err != nil {
		return fmt.Errorf("initial storage access failed: %w", err)
	} else if taskCount == 0 {
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

	if config, err := parseFlags(); err != nil {
		panic(err)
	} else if storage, err := createStorage(ctx, config); err != nil {
		panic(err)
	} else if err := initStorage(ctx, storage); err != nil {
		panic(err)
	} else {
		defer storage.Close()

		log.Info("Listening on " + config.Addr)
		log.Error("listen and serve failed", web.Serve(config.Addr, storage))
	}
}
