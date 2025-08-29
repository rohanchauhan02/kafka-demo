#!/bin/bash
set -euo pipefail

# ========================
# Kafka Topic Setup Script
# ========================

# Configurable variables
BROKER="${BROKER:-localhost:9092}"   # Kafka broker address
TOPICS=("orders" "payments" "notifications")  # Topics to create
PARTITIONS="${PARTITIONS:-3}"        # Default partitions
REPLICATION_FACTOR="${REPLICATION_FACTOR:-2}" # Default replication factor
KAFKA_CONTAINER="${KAFKA_CONTAINER:-kafka}"   # Kafka docker container name

echo "🔹 Kafka Topic Provisioning Started"
echo "Broker: $BROKER"
echo "Partitions: $PARTITIONS, Replication Factor: $REPLICATION_FACTOR"
echo "Topics: ${TOPICS[*]}"

for topic in "${TOPICS[@]}"; do
  echo "➡️ Checking topic: $topic"

  if docker exec "$KAFKA_CONTAINER" kafka-topics \
      --bootstrap-server "$BROKER" --list | grep -q "^$topic$"; then
    echo "✅ Topic '$topic' already exists, skipping..."
  else
    echo "⚙️ Creating topic: $topic"
    docker exec "$KAFKA_CONTAINER" kafka-topics --create \
      --topic "$topic" \
      --partitions "$PARTITIONS" \
      --replication-factor "$REPLICATION_FACTOR" \
      --bootstrap-server "$BROKER"

    echo "✅ Topic '$topic' created successfully."
  fi
done

echo "🎉 Kafka Topic Provisioning Completed"




#!/bin/bash
set -euo pipefail

KAFKA_CONTAINER="kafka"
BROKER="localhost:9092"

case "${1:-}" in
  start)
    echo "🚀 Starting ZooKeeper & Kafka..."
    docker-compose up -d
    ;;

  stop)
    echo "🛑 Stopping Kafka & ZooKeeper..."
    docker-compose down
    ;;

  create-topic)
    topic="${2:-foobar}"
    echo "⚙️ Creating topic: $topic"
    docker exec "$KAFKA_CONTAINER" kafka-topics \
      --create --bootstrap-server $BROKER \
      --replication-factor 1 --partitions 1 --topic "$topic"
    ;;

  list-topics)
    echo "📋 Available topics:"
    docker exec "$KAFKA_CONTAINER" kafka-topics \
      --bootstrap-server $BROKER --list
    ;;

  producer)
    topic="${2:-foobar}"
    echo "✍️ Starting producer for topic: $topic"
    docker exec -it "$KAFKA_CONTAINER" kafka-console-producer \
      --broker-list $BROKER --topic "$topic"
    ;;

  consumer)
    topic="${2:-foobar}"
    echo "👀 Starting consumer for topic: $topic"
    docker exec -it "$KAFKA_CONTAINER" kafka-console-consumer \
      --bootstrap-server $BROKER --topic "$topic" --from-beginning
    ;;

  *)
    echo "Usage: $0 {start|stop|create-topic <name>|list-topics|producer <name>|consumer <name>}"
    exit 1
    ;;
esac
