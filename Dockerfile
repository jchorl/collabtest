FROM golang:latest
ADD . /go/src/github.com/jchorl/collabtest
WORKDIR /go/src/github.com/jchorl/collabtest
RUN go get ./...
ENTRYPOINT go run server.go

