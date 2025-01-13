VERSION := $(shell git describe --tags --always --dirty)

all: build

version:
	@echo $(VERSION)

clean:
	go clean -i ./...

test:
	go test -cover ./...

build: test
	go build ./...

update:
	go get -u ./...

