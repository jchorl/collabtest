.PHONY: all build clean run network db

all: build clean network db run

build:
	docker build -t collabtest/collabtest .

clean:
	-docker rm -f collabtest collabtestdb 2> /dev/null
	-docker network rm collabtest-network

network:
	docker network create collabtest-network

run:
	docker run -it --rm -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock --name collabtest --network collabtest-network -e POSTGRES_USER=collabtest -e POSTGRES_PASSWORD=collabtest collabtest/collabtest

db: 
	docker run --name collabtestdb --expose=5432 -v $(PWD)/sql:/docker-entrypoint-initdb.d -h collabtestdb --network collabtest-network -e POSTGRES_USER=collabtest -e POSTGRES_PASSWORD=collabtest -d postgres
