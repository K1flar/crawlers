package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/K1flar/crawlers/internal/message_broker/kafka"
	"github.com/K1flar/crawlers/internal/message_broker/messages"
	"github.com/K1flar/crawlers/internal/worker"
	"github.com/jmoiron/sqlx"
	dotenv "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	postgresDSN = ("PG_DSN")

	kafkaHost           = ("KAFKA_HOST")
	kafkaPort           = ("KAFKA_PORT")
	tasksToProcessTopic = ("KAFKA_TASKS_TOPIC")
)

type cmd func(ctx context.Context)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if len(os.Args) != 2 {
		log.Error("no cli slug")
		os.Exit(1)
	}

	cliSlug := os.Args[1]

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

	kafkaBrokers := []string{
		fmt.Sprintf("%s:%s", os.Getenv(kafkaHost), os.Getenv(kafkaPort)),
	}

	producer := kafka.NewProducer[messages.TaskToProcessMessage](kafkaBrokers, os.Getenv(tasksToProcessTopic))
	consumer := kafka.NewConsumer[messages.TaskToProcessMessage](kafkaBrokers, os.Getenv(tasksToProcessTopic))

	cron := worker.NewWithPeriod(func(ctx context.Context) {
		err := producer.Produce(ctx, messages.TaskToProcessMessage{ID: rand.Int63()})
		if err != nil {
			log.Error(err.Error())
		}
	}, time.Second)

	worker := worker.New(func(ctx context.Context) {
		msg, err := consumer.Consume(ctx)
		if err != nil {
			log.Error(err.Error())
		}

		fmt.Printf("message: %v\n", msg)
	})

	cmds := map[string]cmd{
		"first": func(ctx context.Context) {
			cron.Run(ctx)
			cron.Wait()
		},
		"second": func(ctx context.Context) {
			worker.Run(ctx)
			worker.Wait()
		},
	}

	cmd, ok := cmds[cliSlug]
	if !ok {
		log.Error("unknown cmd")
		os.Exit(1)
	}

	cmd(ctx)

	producer.Close()
	consumer.Close()
}
