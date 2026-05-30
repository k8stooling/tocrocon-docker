default: help

build:
	docker buildx build --platform linux/amd64 -t tocrocon-docker:1.1.0 --load .

test:
	go test -v ./...

clean:
	rm -rf bin/*

help:
	@echo 'Usage: make (build | test | clean)'