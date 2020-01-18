TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)
export TAG

GO=go

.PHONY: all
all: test build pack

.PHONY: test
test:
	$(GO) get -t ./...
	$(GO) test ./...

.PHONY: build
build:
	$(GO) get -t ./...
	$(GO) build -ldflags "-X main.version=$(TAG)" -o server ./examples/simple/server/.

.PHONY: install
install:
	$(GO) get -t ./...
	$(GO) install  -ldflags "-X main.version=$(TAG)" ./examples/simple/server/.

.PHONY: pack
pack: build
	docker build -t simple-server:${TAG} .

.PHONY: clean
clean:
	$(GO) clean
	@rm -f server
