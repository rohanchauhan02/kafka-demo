#!/bin/bash
for topic in orders payments notifications; do
  docker exec kafka kafka-topics --create \
    --topic $topic \
    --partitions 2 \
    --replication-factor 1 \
    --bootstrap-server localhost:9092
done
