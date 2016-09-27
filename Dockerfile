FROM golang:latest

RUN apt-get update && \
        apt-get install -y inotify-tools

RUN wget https://get.docker.com/builds/Linux/x86_64/docker-latest.tgz && \
    tar zxf docker-latest.tgz && \
    mv docker/docker /usr/bin && \
    rm -rf docker && \
    rm docker-latest.tgz

ADD . /go/src/github.com/jchorl/collabtest
WORKDIR /go/src/github.com/jchorl/collabtest

RUN go-wrapper download
RUN go-wrapper install

CMD ["./start.sh"]
