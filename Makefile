BUILD_COMMIT := $(shell git log --format="%H" -n 1)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%S')
FLAGS = GOOS=linux GOARCH=amd64 CGO_ENABLED=0

PROJECT = github.com/stepan2volkov/urlshortener
CMD:= $(PROJECT)/cmd/urlshortener

check:
	golangci-lint run -c golangci-lint.yaml

test:
	go test ./...

generate:
	go generate ./...

.PHONY: build
build:
	mkdir -p build
	$(FLAGS) go build -a -tags netgo -ldflags="\
		-w -extldflags '-static'\
		-X '$(PROJECT)/app/config.BuildCommit=$(BUILD_COMMIT)'\
		-X '${PROJECT}/app/config.BuildTime=${BUILD_TIME}'"\
		-o build $(CMD)
		

clean:
	rm -rf build

run:
	go run cmd/urlshortener/main.go -config=./config/example.yaml
	
up:
	docker-compose up -d
	sh ./scripts/open-page.sh

down:
	docker-compose down
	docker image rm stepan2volkov/urlshortener:v1.0.0
