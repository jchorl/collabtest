all:
	docker build -t collabtest/collabtest . && docker run -it --rm -p 8080:8080 collabtest/collabtest
