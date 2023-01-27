.PHONY:
.SILENT:
.DEFAULT_GOAL := run

build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o ./.bin/vpn-wg ./cmd/app/main.go

run: build
	docker-compose up --remove-orphans vpn-wg

rebuild:
	docker-compose up -d --no-deps --build