all:clean docker_up
	go build -o build/app ./cmd/main/main.go && ./build/app
clean:
	-rm -r build logs
docker_up:
	docker compose up -d
	sleep 1
consumer:
	go build -o build/consumer ./cmd/kafka-consumer/main.go && ./build/consumer

test:
	go test -v ./tests
