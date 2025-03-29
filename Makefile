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
	go build -o bin/server cmd/server/main.go
	
run:
	go run cmd/server/main.go