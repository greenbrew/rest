TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)
export TAG

all: test build pack
test:
	go test ./...

build:
	go get -t ./...
	go build -ldflags "-X main.version=$(TAG)" -o server ./examples/simple/server/.

install:
	go get -t ./...
	go install  -ldflags "-X main.version=$(TAG)" ./examples/simple/server/.

pack: build
	docker build -t simple-server:${TAG} .

clean:
	go clean
	rm -f server
