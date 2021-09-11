BUILD_COMMIT := $(shell git log --format="%H" -n 1)
BUILD_TIME := $(shell date -u '+%Y/%m/%d %H:%M:%S')

PROJECT = github.com/stepan2volkov/urlshortener
CMD:= $(PROJECT)/cmd/urlshortener

check:
	golangci-lint run -c golangci-lint.yaml

test:
	go test ./...

.PHONY: build
build:
	mkdir -p build
	go build  -ldflags="\
		-X '$(PROJECT)/app/config.BuildCommit=$(BUILD_COMMIT)'\
		-X '${PROJECT}/app/config.BuildTime=${BUILD_TIME}'"\
		-o build $(CMD)
		

clean:
	rm -rf build
