.PHONY: all build clean run network db dev run-dev ui ui-dev

DOMAIN ?= localhost
PORT ?= 4443
DB_DOMAIN ?= collabtestdb
DB_PORT ?= 5432
DB_USER ?= collabtest
DB_PASSWORD ?= collabtest

all: clean network db ui build run
dev: clean network db ui build run-dev

build:
	docker build -t collabtest/collabtest .

clean:
	-docker rm -f collabtest collabtestdb 2> /dev/null
	-docker network rm collabtest-network

network:
	docker network create collabtest-network

ui:
	docker run --rm --name collabtestui -it -v $(PWD)/ui:/usr/src/app -w /usr/src/app node:latest /bin/bash -c "npm install; npm run build"

ui-dev:
	docker run --rm --name collabtestui -it --network collabtest-network -p 3000:3000 -v $(PWD)/ui:/usr/src/app -w /usr/src/app node:latest /bin/bash -c "npm install; HTTPS=true npm start"

run:
	docker run \
		-p $(PORT):$(PORT) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		--name collabtest \
		--network collabtest-network \
		-e DOMAIN=$(DOMAIN) \
		-e PORT=$(PORT) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DOMAIN=$(DB_DOMAIN) \
		-e POSTGRES_PORT=$(DB_PORT) \
		collabtest/collabtest
run-dev:
	docker run -it --rm \
		-p $(PORT):$(PORT) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PWD):/go/src/github.com/jchorl/collabtest \
		-h collabtest \
		--name collabtest \
		--network collabtest-network \
		-e DEV=1 \
		-e DOMAIN=$(DOMAIN) \
		-e PORT=$(PORT) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DOMAIN=$(DB_DOMAIN) \
		-e POSTGRES_PORT=$(DB_PORT) \
		collabtest/collabtest

db: 
	docker run --name collabtestdb --expose=$(DB_PORT) -v $(PWD)/sql:/docker-entrypoint-initdb.d -h collabtestdb --network collabtest-network -e POSTGRES_USER=collabtest -e POSTGRES_PASSWORD=collabtest -d postgres
	sleep 5
