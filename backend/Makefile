.PHONY: run build test lint wire migrate air clean fresh seed

APP_NAME := koala-exam
MAIN_PATH := cmd/hertz
BIN_PATH := bin/${APP_NAME}

CONFIG_FILE ?= configs/config.yaml

run:
	go run ./${MAIN_PATH}

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ${BIN_PATH} ./${MAIN_PATH}

air:
	air -c .air.toml

test:
	go test ./... -v -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

# 一键建库 + 建表 + 种子数据
fresh:
	go run ./cmd/migrate -op fresh

# 仅建表（不删）
migrate-up:
	go run ./cmd/migrate -op up

# 仅种子数据
seed:
	go run ./cmd/migrate -op seed

# 清空 + 建表
reset:
	go run ./cmd/migrate -op reset

# 删除所有表
migrate-down:
	go run ./cmd/migrate -op down

clean:
	rm -rf bin/ tmp/ coverage.out coverage.html logs/

docker-build:
	docker build -t koala-exam-backend .

docker-run:
	docker run --rm -p 8080:8080 --env-file configs/.env.production koala-exam-backend

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

deps:
	go install github.com/air-verse/air@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
