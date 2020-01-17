FROM golang:1.11 as builder
COPY . ./src/github.com/greenbrew/rest
WORKDIR ./src/github.com/greenbrew/rest
RUN go get -t -v ./...
RUN go test ./...
RUN version=$(git rev-list HEAD --max-count=1 --abbrev-commit)
RUN go install -ldflags "-X main.version=$(version)" ./...

# Take the simple server into a docker image
FROM ubuntu:18.04
WORKDIR /simple/
COPY --from=builder /go/bin/server /simple/server
EXPOSE 8443
ENTRYPOINT ./server
