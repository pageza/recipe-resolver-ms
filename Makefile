.PHONY: build run test clean

build:
	go build -o resolver-microservice

run: build
	./resolver-microservice

test:
	go test -v ./...

clean:
	rm -f resolver-microservice 