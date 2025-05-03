package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/K1flar/crawlers/internal/actions/consume_tasks_to_process"
	produce_tasks_to_process_action "github.com/K1flar/crawlers/internal/actions/produce_tasks_to_process"
	"github.com/K1flar/crawlers/internal/gates/searx"
	"github.com/K1flar/crawlers/internal/gates/web_scraper"
	"github.com/K1flar/crawlers/internal/http_client"
	"github.com/K1flar/crawlers/internal/message_broker/kafka"
	"github.com/K1flar/crawlers/internal/message_broker/messages"
	"github.com/K1flar/crawlers/internal/services/crawler"
	"github.com/K1flar/crawlers/internal/storage/tasks"
	"github.com/K1flar/crawlers/internal/stories/process_task"
	"github.com/K1flar/crawlers/internal/stories/produce_tasks_to_process"
	"github.com/K1flar/crawlers/internal/worker"
	"github.com/jmoiron/sqlx"
	dotenv "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	postgresDSN = "PG_DSN"

	kafkaHost           = "KAFKA_HOST"
	kafkaPort           = "KAFKA_PORT"
	tasksToProcessTopic = "KAFKA_TASKS_TOPIC"
	consumerGroupID     = "consumer-group"

	searxHost = "SEARX_HOST"
	searxPort = "SEARX_PORT"

	maxCountCrawlers    = "MAX_COUNT_CRAWLERS"
	defaulCountCrawlers = 10

	cronTasksToProcessPeriod    = "CRON_TASKS_TO_PROCESS_PRODUCER_PERIOD"
	defaultTasksToProcessPeriod = time.Hour * 24
)

type cmd func(ctx context.Context)

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if len(os.Args) != 2 {
		log.Error("no cli slug")
		os.Exit(1)
	}

	cliSlug := os.Args[1]

	err = dotenv.Load()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	maxCountCrawlersInt, err := strconv.Atoi(os.Getenv(maxCountCrawlers))
	if err != nil {
		log.Warn(fmt.Sprintf("failed to parse max count crawlers: [%s]", os.Getenv(maxCountCrawlers)))

		maxCountCrawlersInt = defaulCountCrawlers
	}

	tasksToProcessPeriod, err := time.ParseDuration(cronTasksToProcessPeriod)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to parse cron tasks to process producer period: [%s]", os.Getenv(cronTasksToProcessPeriod)))

		tasksToProcessPeriod = defaultTasksToProcessPeriod
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

	kafkaBrokers := []string{
		fmt.Sprintf("%s:%s", os.Getenv(kafkaHost), os.Getenv(kafkaPort)),
	}

	producer := kafka.NewProducer[messages.TaskToProcessMessage](kafkaBrokers, os.Getenv(tasksToProcessTopic))
	consumer := kafka.NewConsumer[messages.TaskToProcessMessage](kafkaBrokers, os.Getenv(tasksToProcessTopic))

	// Clients
	searxClient := http_client.New(
		http_client.WithBaseURL(os.Getenv(searxHost) + ":" + os.Getenv(searxPort)),
	)

	// Storages
	tasksStorage := tasks.NewStorage(db)
	// sourcesStorage := sources.NewStorage(db)

	// Gates
	sxGate := searx.NewGate(log, searxClient)
	webScraperGate := web_scraper.NewGate()

	// Services
	crawler := crawler.New(log, sxGate, webScraperGate)

	// Stories
	produceAllActiveTasksToProcessStory := produce_tasks_to_process.NewStory(tasksStorage, producer)
	processTaskStory := process_task.NewStory(log, tasksStorage, crawler)

	// Actions
	tasksToProcessProducer := produce_tasks_to_process_action.NewAction(log, produceAllActiveTasksToProcessStory)
	tasksToProcessConsumer := consume_tasks_to_process.NewAction(log, consumer, processTaskStory, maxCountCrawlersInt)

	cmds := map[string]cmd{
		"tasks-to-process-producer": worker.NewWithPeriod(tasksToProcessProducer.Run, tasksToProcessPeriod).Run,
		"tasks-to-process-consumer": worker.New(tasksToProcessConsumer.Run).Run,
	}

	cmd, ok := cmds[cliSlug]
	if !ok {
		log.Error("unknown cmd")
		os.Exit(1)
	}

	log.Info(fmt.Sprintf("start [%s] process", cliSlug))
	cmd(ctx)

	producer.Close()
	consumer.Close()
}
