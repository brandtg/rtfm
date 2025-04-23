APP_NAME := rtfm

.PHONY: build clean install fmt test

build:
	go build -o bin/$(APP_NAME) .

install:
	go install .

clean:
	rm -rf bin/

fmt:
	go fmt ./...

license:
	go run github.com/google/addlicense@latest -c "Greg Brandt" -l apache .

test:
	go test ./...
