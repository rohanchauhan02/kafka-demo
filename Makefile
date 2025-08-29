
SHELL := /bin/bash
BROKER := localhost:9092
TOPIC ?= foobar

setup:
	chmod +x scripts/create-topics.sh
	./scripts/create-topics.sh

start-producer:
	go run cmd/producer/main.go

start-consumer1:
	go run cmd/consumer-group1/main.go

start-consumer2:
	go run cmd/consumer-group2/main.go

run: start-producer start-consumer1 start-consumer2

start:
	@echo "🚀 Starting ZooKeeper & Kafka..."
	docker-compose up -d

stop:
	@echo "🛑 Stopping ZooKeeper & Kafka..."
	docker-compose down

restart: stop start

logs:
	@echo "📜 Showing Kafka logs..."
	docker logs -f kafka

create-topic:
	@echo "⚙️ Creating topic: $(TOPIC)"
	docker exec kafka kafka-topics \
		--create --bootstrap-server $(BROKER) \
		--replication-factor 1 --partitions 1 --topic $(TOPIC) || true

list-topics:
	@echo "📋 Listing topics..."
	docker exec kafka kafka-topics \
		--bootstrap-server $(BROKER) --list

producer:
	@echo "✍️ Starting producer for topic: $(TOPIC)"
	docker exec -it kafka kafka-console-producer \
		--broker-list $(BROKER) --topic $(TOPIC)

consumer:
	@echo "👀 Starting consumer for topic: $(TOPIC)"
	docker exec -it kafka kafka-console-consumer \
		--bootstrap-server $(BROKER) --topic $(TOPIC) --from-beginning

.PHONY: run start stop restart logs create-topic list-topics producer consumer start setup producer consumer1 consumer2
