.PHONY: start setup producer consumer1 consumer2

start:
	docker compose up -d

setup:
	chmod +x scripts/create-topics.sh
	./scripts/create-topics.sh

producer:
	go run cmd/producer/main.go

consumer1:
	go run cmd/consumer-group1/main.go

consumer2:
	go run cmd/consumer-group2/main.go
