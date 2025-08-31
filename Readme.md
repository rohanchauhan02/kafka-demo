# ⚡ Kafka Multi-Topic Multi-Consumer System (Go + Sarama)

A **production-ready Kafka architecture** built in **Go**, demonstrating a real-world **multi-topic, multi-consumer group setup** using the [Sarama](https://github.com/IBM/sarama) Kafka client.

---

## 📊 Architecture Overview

```
             ┌────────────┐
             │  Producer  │
             └────┬───────┘
                  │
       ┌──────────▼───────────┐
       │      Kafka Topics    │
       │                      │
       │  ┌───────────────┐   │
       │  │   orders      │◄──┼────────────┐
       │  └───────────────┘   │            │
       │  ┌───────────────┐   │            │
       │  │   payments     │◄─┼─────┐      │
       │  └───────────────┘   │     │      │
       │  ┌───────────────┐   │     │      │
       │  │ notifications │◄──┼─────┘      │
       │  └───────────────┘   │            │
       └──────────────────────┘            │
                  ▲                        │
         ┌────────┴──────┐         ┌───────▼────────┐
         │ ConsumerGroup1│         │ ConsumerGroup2 │
         │ orders,       │         │ payments,      │
         │ payments      │         │ notifications  │
         └───────────────┘         └────────────────┘
```

---

## 🧠 Kafka Concepts Simplified

### 🔹 Topics

Logical channels where events (messages) are published.
This project uses:

* `orders`
* `payments`
* `notifications`

Each topic has **2 partitions** for parallel consumption.

### 🔹 Partitions

Partitions let Kafka **scale horizontally** and **preserve order** within a partition.
We use **hash-based partitioning** by event ID → ensures ordering per event.

### 🔹 Producer

* Sends structured JSON events into Kafka
* Uses **asynchronous delivery** (high throughput)
* Built with `sarama.AsyncProducer`

### 🔹 Consumer Groups

A **consumer group** is a set of consumers that share work.

| Group             | Topics Consumed             |
| ----------------- | --------------------------- |
| `service-group-1` | `orders`, `payments`        |
| `service-group-2` | `payments`, `notifications` |

Each group has independent offsets → decoupled progress.

### 🔹 Offsets

Offsets = last-read position in a partition.
They enable replay, failure recovery, and consumer independence.

---

## 🏗 Project Structure

```
kafka-system/
├── cmd/
│   ├── producer/           # Producer logic
│   ├── consumer-group1/    # Group 1: orders + payments
│   └── consumer-group2/    # Group 2: payments + notifications
├── internal/
│   ├── kafka/              # Shared Kafka setup
│   └── types/              # Event & Stats types
├── scripts/
│   └── create-topics.sh    # Topic creation script
├── docker-compose.yml      # Kafka + Zookeeper setup
├── Makefile                # Developer shortcuts
└── README.md               # You're here 🚀
```

---

## ⚙️ How to Run Locally

### ✅ 1. Start Kafka & Zookeeper

```bash
make start
```

Runs Kafka (`localhost:9092`) and Zookeeper (`localhost:2181`).

### ✅ 2. Create Topics

```bash
make setup
```

Creates: `orders`, `payments`, `notifications` with **2 partitions each**.

### ✅ 3. Run Services in Separate Terminals

**Terminal 1 – Producer**

```bash
make producer
```

**Terminal 2 – Consumer Group 1**

```bash
make consumer1
```

**Terminal 3 – Consumer Group 2**

```bash
make consumer2
```

---

## 🔍 Sample Event Output

```json
{
  "type": "order_created",
  "id": "evt-1721574564890993100",
  "data": "data-128",
  "timestamp": "2025-07-22T12:15:00Z"
}
```

---

## 📜 Makefile Commands

| Command          | Purpose                       |
| ---------------- | ----------------------------- |
| `make start`     | Start Kafka & Zookeeper       |
| `make setup`     | Create topics (`orders` etc.) |
| `make producer`  | Run producer                  |
| `make consumer1` | Run Consumer Group 1          |
| `make consumer2` | Run Consumer Group 2          |
| `make stop`      | Stop Kafka & Zookeeper        |
| `make logs`      | Show Kafka logs               |

---

## 📦 Dependencies

* Go ≥ 1.18
* [Sarama](https://github.com/IBM/sarama) Kafka client
* Docker & Docker Compose

---

## 💡 Real-World Use Cases

* Event-driven microservices → orders, payments, notifications
* Log aggregation → centralizing service logs
* ETL pipelines → stream-transform-load architecture

---

## 🔒 Graceful Shutdowns

Consumers trap `SIGINT` & `SIGTERM`, print summaries, and exit cleanly.

Example:

```
[orders][0] order_created: evt-1721574564890993100
[orders][1] order_updated: evt-1721574564890993115

Consumption Summary:
 orders[0]: 12
 orders[1]: 9
 orders TOTAL: 21
```

---

## 🧪 Extensibility Ideas

* [ ] Add Prometheus metrics + Grafana dashboards
* [ ] REST API gateway for publishing events
* [ ] Schema Registry (Avro/Protobuf)
* [ ] Retry & Dead Letter Queues (DLQs)

---

## 🧼 Cleanup

```bash
docker-compose down        # Stop containers
docker-compose down -v     # Stop + remove volumes
```

---

## 🤝 Credits

Built with ❤️ using:

* [Apache Kafka](https://kafka.apache.org/)
* [Sarama](https://github.com/IBM/sarama)
* Go + Docker

---
