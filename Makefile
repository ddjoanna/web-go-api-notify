#!make

include .env
export $(shell sed 's/=.*//' .env)

DOCKER_COMPOSE_FILE ?= docker-compose.yml

#========================#
#== DEVELOPMENT ==#
#========================#

up:
	docker compose -f ${DOCKER_COMPOSE_FILE} up -d --remove-orphans

down:
	docker compose -f ${DOCKER_COMPOSE_FILE} down

install:
	git submodule update --init --recursive && \
	go mod download && \
	go mod tidy

grpcui:
	grpcui -plaintext localhost:${GRPC_PORT}

#========================#
#== BUILD & RUN ==#
#========================#
build:
	go build -o bin/server cmd/server/main.go && \
	go build -o bin/worker cmd/worker/main.go && \
	go build -o bin/cronjob cmd/cronjob/main.go
	
run:
	go run cmd/server/main.go

run-worker:
	go run cmd/worker/main.go

run-cronjob:
	go run cmd/cronjob/main.go

#========================#
#== KAFKA ==#
#========================#

kafka-create-topics:
	docker compose -f ${DOCKER_COMPOSE_FILE} exec kafka kafka-topics.sh --create --topic 'notify-sms' --bootstrap-server kafka:9092 --replication-factor 1 --partitions 24
	docker compose -f ${DOCKER_COMPOSE_FILE} exec kafka kafka-topics.sh --create --topic 'notify-mail' --bootstrap-server kafka:9092 --replication-factor 1 --partitions 24

kafka-clear-topics:
	docker compose -f ${DOCKER_COMPOSE_FILE} exec kafka kafka-topics.sh --delete --topic '$(TOPIC)' --bootstrap-server kafka:9092

kafka-list-topics:
	docker compose -f ${DOCKER_COMPOSE_FILE} exec kafka kafka-topics.sh --list --bootstrap-server kafka:9092

kafka-console-consumer:
	docker compose -f ${DOCKER_COMPOSE_FILE} exec kafka kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic '$(TOPIC)' --from-beginning --max-messages 10