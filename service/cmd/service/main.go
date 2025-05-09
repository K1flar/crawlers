package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	api_activate_task "github.com/K1flar/crawlers/internal/handlers/activate_task"
	api_create_task "github.com/K1flar/crawlers/internal/handlers/create_task"
	api_get_protocol "github.com/K1flar/crawlers/internal/handlers/get_protocol"
	api_get_sources "github.com/K1flar/crawlers/internal/handlers/get_sources"
	api_get_task "github.com/K1flar/crawlers/internal/handlers/get_task"
	api_get_task_status "github.com/K1flar/crawlers/internal/handlers/get_task_status"
	api_get_tasks "github.com/K1flar/crawlers/internal/handlers/get_tasks"
	api_stop_task "github.com/K1flar/crawlers/internal/handlers/stop_task"
	api_update_task "github.com/K1flar/crawlers/internal/handlers/update_task"
	"github.com/K1flar/crawlers/internal/message_broker/kafka"
	"github.com/K1flar/crawlers/internal/message_broker/messages"
	"github.com/K1flar/crawlers/internal/middlewares/cors"
	"github.com/K1flar/crawlers/internal/storage/launches"
	"github.com/K1flar/crawlers/internal/storage/sources"
	"github.com/K1flar/crawlers/internal/storage/tasks"
	"github.com/K1flar/crawlers/internal/stories/create_task"
	"github.com/jmoiron/sqlx"
	dotenv "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	serviceHost = "SERVICE_HOST"
	servicePort = "SERVICE_PORT"

	postgresDSN = "PG_DSN"

	kafkaHost           = "KAFKA_HOST"
	kafkaPort           = "KAFKA_PORT"
	tasksToProcessTopic = "KAFKA_TASKS_TOPIC"

	searxHost = "SEARX_HOST"
	searxPort = "SEARX_PORT"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := dotenv.Load()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", os.Getenv(postgresDSN))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Storage
	tasksStorage := tasks.NewStorage(db)
	sourcesStorage := sources.NewStorage(db)
	launchesStorage := launches.NewStorage(db)

	kafkaBrokers := []string{
		fmt.Sprintf("%s:%s", os.Getenv(kafkaHost), os.Getenv(kafkaPort)),
	}

	producerTasksToProcess := kafka.NewProducer[messages.TaskToProcessMessage](kafkaBrokers, os.Getenv(tasksToProcessTopic))

	createTaskStory := create_task.NewStory(log, tasksStorage, producerTasksToProcess)

	mux := http.NewServeMux()

	corsMW := cors.New()

	mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	mux.Handle("POST /create-task", corsMW(http.HandlerFunc(api_create_task.New(log, createTaskStory).Handle)))
	mux.Handle("POST /get-task", corsMW(http.HandlerFunc(api_get_task.New(log, tasksStorage, launchesStorage).Handle)))
	mux.Handle("POST /get-task-status", corsMW(http.HandlerFunc(api_get_task_status.New(log, tasksStorage).Handle)))
	mux.Handle("POST /get-sources", corsMW(http.HandlerFunc(api_get_sources.New(log, sourcesStorage).Handle)))
	mux.Handle("POST /stop-task", corsMW(http.HandlerFunc(api_stop_task.New(log, tasksStorage).Handle)))
	mux.Handle("POST /activate-task", corsMW(http.HandlerFunc(api_activate_task.New(log, tasksStorage, producerTasksToProcess).Handle)))
	mux.Handle("POST /update-task", corsMW(http.HandlerFunc(api_update_task.New(log, tasksStorage).Handle)))
	mux.Handle("POST /get-tasks", corsMW(http.HandlerFunc(api_get_tasks.New(log, tasksStorage).Handle)))
	mux.Handle("POST /get-protocol", corsMW(http.HandlerFunc(api_get_protocol.New(log, sourcesStorage).Handle)))

	log.Info(fmt.Sprintf("Starting server on %s:%s", os.Getenv(serviceHost), os.Getenv(servicePort)))
	if err := http.ListenAndServe(os.Getenv(serviceHost)+":"+os.Getenv(servicePort), mux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
