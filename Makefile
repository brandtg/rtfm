APP_NAME := rtfm

.PHONY: build run clean install

build:
	go build -o bin/$(APP_NAME) .

run:
	go run main.go

install:
	go install .

clean:
	rm -rf bin/

fmt:
	go fmt ./...
