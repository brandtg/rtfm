APP_NAME=rtfm
CMD_DIR=./cmd/$(APP_NAME)

.PHONY: build run test fmt lint clean

build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
