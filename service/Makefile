PG_HOST = 127.0.0.1
PG_PORT = 5432
PG_NAME = crawler_tasks_db
PG_USER = postgres
PG_PASSWORD = postgres

SERVICE_HOST = 127.0.0.1
SERVICE_PORT = 8080

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
	@echo SERVICE_HOST=$(SERVICE_HOST) >> .env 
	@echo SERVICE_PORT=$(SERVICE_PORT) >> .env 
	@echo Environment variables have been successfully created

.PHONY: clean-env
clean-env:
	@$(RM) -f .env

.PHONY: db-start
db-start:
	@docker-compose up db -d
	@echo PostgreSQL is running 

.PHONY: workers-start
workers-start:
	@docker-compose --env-file .env up workers -d
	@echo Workers is running

.PHONY: service-start
service-start:
	@docker-compose --env-file .env up service -d
	@echo Service is running

.PHONY: stop
stop:
	@docker-compose down 
	@make clean-env
	@echo The app is stopped

.PHONY: service-debug
service-debug: env db-start

.PHONY: service-deploy
service-deploy: env db-start service-start workers-start