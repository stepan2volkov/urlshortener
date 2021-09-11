BUILD_COMMIT := $(shell git log --format="%H" -n 1)
PROJECT = github.com/stepan2volkov/urlshortener
CMD:= PROJECT + /cmd/urlshortener

check:
	golangci-lint run -c golangci-lint.yaml

test:
	go test ./...

.PHONY: build
build:
	mkdir -p build
	go build -o build -ldflags="-X '$(PROJECT)/app/config.BuildCommit=$(BUILD_COMMIT)'" $(CMD)
		

clean:
	rm -rf build
