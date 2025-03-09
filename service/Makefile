PG_HOST = 127.0.0.1
PG_PORT = 5432
PG_NAME = crawler_tasks_db
PG_USER = postgres
PG_PASSWORD = postgres
PG_DSN = postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_NAME)?sslmode=disable
MIGRATION_DIR = db/migrations

SERVICE_HOST = 127.0.0.1
SERVICE_PORT = 8080

SEARX_HOST = 127.0.0.1
SEARX_PORT = 8888

KAFKA_HOST = 127.0.0.1
KAFKA_PORT = 9092
KAFKA_TASKS_TOPIC = tasks-to-create

MIGRATE = docker run \
	-v ${CURDIR}/$(MIGRATION_DIR):/migrations \
	--network host \
	migrate/migrate \
    -path=/migrations

RM = rm -f
ifeq ($(OS),Windows_NT)
    RM = del /Q
endif

.PHONY: env
env:
	@echo PG_HOST=$(PG_HOST) > .env 
	@echo PG_PORT=$(PG_PORT) >> .env 
	@echo PG_NAME=$(PG_NAME) >> .env 
	@echo PG_USER=$(PG_USER) >> .env 
	@echo PG_PASSWORD=$(PG_PASSWORD) >> .env 
	@echo PG_DSN=$(PG_DSN) >> .env
	@echo SERVICE_HOST=$(SERVICE_HOST) >> .env 
	@echo SERVICE_PORT=$(SERVICE_PORT) >> .env 
	@echo SEARX_HOST=$(SEARX_HOST) >> .env 
	@echo SEARX_PORT=$(SEARX_PORT) >> .env 
	@echo KAFKA_HOST=$(KAFKA_HOST) >> .env 
	@echo KAFKA_PORT=$(KAFKA_PORT) >> .env 
	@echo KAFKA_TASKS_TOPIC=$(KAFKA_TASKS_TOPIC) >> .env
	@echo Environment variables have been successfully created

.PHONY: clean-env
clean-env:
	@$(RM) -f .env

.PHONY: db-start
db-start:
	@docker-compose up db -d
	@echo PostgreSQL is running 

.PHONY: migrate-up
migrate-up:
	@$(MIGRATE) -database $(PG_DSN) up
	@echo Migrations applied

.PHONY: migrate-down
migrate-down:
	@$(MIGRATE) -database $(PG_DSN) down --all
	@echo Migrations applied

.PHONY: workers-start
workers-start:
	@docker-compose --env-file .env up workers -d
	@echo Workers is running

.PHONY: service-start
service-start:
	@docker-compose --env-file .env up service -d
	@echo Service is running

.PHONY: searx-start
searx-start:
	@docker-compose up searx -d
	@echo Searx is running

.PHONY: kafka-start
kafka-start: 
	@docker-compose up kafka -d
	@echo Kafka cluster is running

.PHONY: create-tasks-topic
create-tasks-topic: 
	@docker exec -it kafka \
	/opt/kafka/bin/kafka-topics.sh --create \
	--topic $(KAFKA_TASKS_TOPIC) \
	--bootstrap-server localhost:9092 \
	--partitions 1 \
	--replication-factor 1
	@echo Topic for start tasks is created

.PHONY: stop
stop:
	@docker-compose down 
	@make clean-env
	@echo The app is stopped

.PHONY: service-debug
service-debug: env db-start migrate-up searx-start kafka-start create-tasks-topic

.PHONY: service-deploy
service-deploy: env db-start migrate-up service-start workers-start searx-start kafka-start create-tasks-topic