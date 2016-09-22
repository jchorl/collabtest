.PHONY: all build clean run

all: build clean run

build:
	docker build -t collabtest/collabtest .

clean:
	-docker rm -f collabtest 2> /dev/null || true

run:
	docker run -it --rm -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock --name collabtest collabtest/collabtest


