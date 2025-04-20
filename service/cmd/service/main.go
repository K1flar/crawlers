package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

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

const (
	serviceHost = "SERVICE_HOST"
	servicePort = "SERVICE_PORT"

	postgresDSN = "PG_DSN"

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

	tasks := tasks.NewStorage(db)
	sources := sources.NewStorage(db)

	searxClient := http_client.New(
		http_client.WithBaseURL(os.Getenv(searxHost) + ":" + os.Getenv(searxPort)),
	)
	searxGate := searx.NewGate(log, searxClient)

	webScraperGate := web_scraper.NewGate()

	crawler := crawler.New(log, searxGate, webScraperGate, sources)

	createTaskStory := create_task.NewStory(log, tasks, crawler)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /create-task", api_create_task.New(log, createTaskStory).Handle)

	log.Info(fmt.Sprintf("Starting server on %s:%s", os.Getenv(serviceHost), os.Getenv(servicePort)))
	if err := http.ListenAndServe(os.Getenv(serviceHost)+":"+os.Getenv(servicePort), mux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
