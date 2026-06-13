VERSION := $(shell git describe --tags --always --dirty)

all: build

version:
	@echo $(VERSION)

clean:
	go clean -i ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

cover: test
	go tool cover -func=coverage.out

bench:
	go test -run=^$$ -bench=. -benchmem ./...

build: vet test
	go build ./...

update:
	go get -u ./...
	go mod tidy
