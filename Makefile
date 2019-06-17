
all: build

clean:
	go clean -i ./...

test:
	go test -cover ./...

build: test
	go build ./...

