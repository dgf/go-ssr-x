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

type storageType int

type storageConfig struct {
	storageType storageType
	connection  string
}

type serverConfig struct {
	addr    string
	storage storageConfig
}

const (
	memoryStorage storageType = iota
	fileStorage
	databaseStorage

	defaultAddr       = "0.0.0.0:3000"
	defaultConnStr    = "postgres://task-db-user:my53cr3tpa55w0rd@localhost?sslmode=disable"
	defaultSqliteFile = ".tasks.sqlite"
)

func parseFlags() (serverConfig, error) {
	var addr, connStr, storage string

	flag.StringVar(&addr, "address", defaultAddr, "web server address")
	flag.StringVar(&storage, "storage", "memory", "memory, file or database")
	flag.StringVar(&connStr, "connection", defaultConnStr, "database connection string")
	flag.Parse()

	config := serverConfig{addr: addr, storage: storageConfig{storageType: memoryStorage, connection: connStr}}

	if !slices.Contains([]string{"memory", "file", "database"}, storage) {
		flag.Usage()

		return config, fmt.Errorf("unknown storage type: %s", storage)
	}

	switch storage {
	case "database":
		config.storage.storageType = databaseStorage
	case "file":
		config.storage.storageType = fileStorage
		if connStr == defaultConnStr {
			config.storage.connection = defaultSqliteFile
		}
	}

	return config, nil
}

func createMemoryStorage() (entity.Storage, error) {
	log.Warn("running with in-memory storage, the data will be lost when restarting")

	return entity.NewMemory(), nil
}

func createDatabaseStorage(ctx context.Context, connStr string) (entity.Storage, error) {
	return postgres.NewDatabase(ctx, connStr)
}

func createFileStorage(ctx context.Context, path string) (entity.Storage, error) {
	log.Info("use file storage", "path", path)

	return sqlite3.NewFile(ctx, path)
}

func createStorage(ctx context.Context, config serverConfig) (entity.Storage, error) {
	switch config.storage.storageType {
	case memoryStorage:
		return createMemoryStorage()
	case databaseStorage:
		return createDatabaseStorage(ctx, config.storage.connection)
	case fileStorage:
		return createFileStorage(ctx, config.storage.connection)
	}

	return nil, fmt.Errorf("unknown storage type: %d", config.storage.storageType)
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

	config, err := parseFlags()
	if err != nil {
		panic(err)
	}

	storage, err := createStorage(ctx, config)
	if err != nil {
		panic(err)
	}

	err = initStorage(ctx, storage)
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	log.Info("Listening on " + config.addr)
	log.Error("listen and serve failed", web.Serve(config.addr, storage))
}
