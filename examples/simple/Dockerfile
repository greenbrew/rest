FROM golang:1.11 as builder
COPY . ./
WORKDIR /go/src/github.com/greenbrew/rest/examples/simple
RUN go get -t ./...
RUN go build -a .
