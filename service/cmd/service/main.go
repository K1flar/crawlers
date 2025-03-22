package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/IBM/sarama"
	"github.com/K1flar/crawlers/internal/gates/searx"
	"github.com/K1flar/crawlers/internal/gates/web_scraper"
	api_create_task "github.com/K1flar/crawlers/internal/handlers/create_task"
	"github.com/K1flar/crawlers/internal/http_client"
	"github.com/K1flar/crawlers/internal/services/crawler"
	"github.com/K1flar/crawlers/internal/storage/sources"
	"github.com/K1flar/crawlers/internal/storage/tasks"
	"github.com/K1flar/crawlers/internal/stories/create_task"
	"github.com/jmoiron/sqlx"
	dotenv "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	serviceHost = os.Getenv("SERVICE_HOST")
	servicePort = os.Getenv("SERVICE_PORT")

	postgresDSN = os.Getenv("PG_DSN")

	searxHost = os.Getenv("SEARX_HOST")
	searxPort = os.Getenv("SEARX_PORT")

	kafkaHost           = os.Getenv("KAFKA_HOST")
	kafkaPort           = os.Getenv("KAFKA_PORT")
	tasksToProcessTopic = os.Getenv("KAFKA_TASKS_TOPIC")
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := dotenv.Load()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", postgresDSN)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	tasks := tasks.NewStorage(db)
	sources := sources.NewStorage(db)

	searxClient := http_client.New(
		http_client.WithBaseURL(searxHost + ":" + searxPort),
	)
	searxGate := searx.NewGate(log, searxClient)

	webScraperGate := web_scraper.NewGate()

	crawler := crawler.New(log, searxGate, webScraperGate, sources)

	createTaskStory := create_task.NewStory(log, tasks, crawler)

	kafkaProducerConfig := sarama.NewConfig()
	kafkaProducerConfig.Producer.Return.Successes = true

	_, err = sarama.NewSyncProducer([]string{kafkaHost, kafkaPort}, kafkaProducerConfig)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /create-task", api_create_task.New(log, createTaskStory).Handle)

	log.Info(fmt.Sprintf("Starting server on %s:%s", serviceHost, servicePort))
	if err := http.ListenAndServe(serviceHost+":"+servicePort, mux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
