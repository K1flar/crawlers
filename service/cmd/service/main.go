package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/K1flar/crawlers/internal/gates/searx"
	"github.com/K1flar/crawlers/internal/gates/web_scraper"
	api_create_task "github.com/K1flar/crawlers/internal/handlers/create_task"
	"github.com/K1flar/crawlers/internal/http_client"
	"github.com/K1flar/crawlers/internal/storage/tasks"
	"github.com/K1flar/crawlers/internal/stories/create_task"
	"github.com/jmoiron/sqlx"
	dotenv "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := dotenv.Load()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", os.Getenv("PG_DSN"))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	tasks := tasks.NewStorage(db)
	createTaskStory := create_task.NewStory(log, tasks)

	searxClient := http_client.New(
		http_client.WithBaseURL(os.Getenv("SEARX_HOST") + ":" + os.Getenv("SEARX_PORT")),
	)
	searxGate := searx.NewGate(log, searxClient)
	res, err := searxGate.Search(context.Background(), "golang")
	fmt.Println(res, err)

	webScraperGate := web_scraper.NewGate(log)
	page, err := webScraperGate.GetPage(context.Background(), res[0])
	fmt.Println(page.URLs)
	fmt.Println(err)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /create-task", api_create_task.New(log, createTaskStory).Handle)

	log.Info(fmt.Sprintf("Starting server on %s:%s", os.Getenv("SERVICE_HOST"), os.Getenv("SERVICE_PORT")))
	if err := http.ListenAndServe(os.Getenv("SERVICE_HOST")+":"+os.Getenv("SERVICE_PORT"), mux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
