FROM golang:latest
RUN wget https://get.docker.com/builds/Linux/x86_64/docker-latest.tgz && \
    tar zxf docker-latest.tgz && \
    mv docker/docker /usr/bin && \
    rm -rf docker && \
    rm docker-latest.tgz
ADD . /go/src/github.com/jchorl/collabtest
WORKDIR /go/src/github.com/jchorl/collabtest
RUN go get ./...
ENTRYPOINT go run server.go

