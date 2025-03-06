.PHONY: build run test clean docker-build docker-run

build:
	go build -o resolver-microservice

run: build
	./resolver-microservice

test:
	go test -v ./...

clean:
	rm -f resolver-microservice 

docker-build:
	docker build --network=host -t recipe-resolver-microservice .

docker-run:
	docker run -it --rm -p 3000:3000 recipe-resolver-microservice 