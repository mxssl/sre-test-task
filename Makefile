BINARY_NAME=app
CURRENT_DIR=$(shell pwd)

.PHONY: all build clean lint critic test dep

all: dep build

build:
	go build -o ${BINARY_NAME} -v
	chmod +x app

clean:
	rm -f ${BINARY_NAME}

lint:
	golangci-lint run

critic:
	gocritic check-project ${CURRENT_DIR}

test:
	go test -v ./...

dep:
	dep ensure
