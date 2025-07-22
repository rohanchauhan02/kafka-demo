# Kafka Multi-Topic Multi-Consumer System (Go + Sarama)

A **production-ready Kafka architecture** built in **Go**, demonstrating a real-world multi-topic, multi-consumer group setup using the [Sarama](https://github.com/Shopify/sarama) Kafka client.

---

## 📊 Architecture Overview

```

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

```

---

## 🧠 Kafka Concepts Simplified

### 🔹 Topic

A **topic** is a logical channel in Kafka where events (messages) are published. This project uses:

- `orders`
- `payments`
- `notifications`

Each topic has **2 partitions** to support parallel consumption.

---

### 🔹 Partition

Partitions allow Kafka to **scale horizontally** and **maintain message order** within each partition.

✅ We use **hash-based partitioning** by event ID to ensure:

- Same event ID → Same partition → Ordered processing

---

### 🔹 Producer

- Sends structured JSON events to topics
- Uses **asynchronous** delivery with success/error handling
- Chooses a topic and event type randomly (simulated event flow)

✅ Built using Sarama's `AsyncProducer`

---

### 🔹 Consumer Group

A **consumer group** is a set of consumers cooperating to consume a set of topics.

| Group            | Topics Consumed             |
|------------------|-----------------------------|
| `service-group-1`| `orders`, `payments`        |
| `service-group-2`| `payments`, `notifications` |

- Each group maintains **independent offsets**
- Partition assignment is managed via **rebalance strategies**
- Consumers **log summaries** and handle graceful shutdowns

---

### 🔹 Offsets

Kafka tracks the **last-read position (offset)** per partition per consumer group. This enables:

- Replay
- Recovery from failures
- Independent progression

---

## 🏗 Project Structure

```

kafka-system/
├── cmd/                    # Main programs
│   ├── producer/           # Producer logic
│   ├── consumer-group1/    # Group 1: orders + payments
│   └── consumer-group2/    # Group 2: payments + notifications
├── internal/
│   ├── kafka/              # Shared Kafka setup (producer/consumer config)
│   └── types/              # Shared Event & Stats types
├── scripts/
│   └── create-topics.sh    # CLI topic creation script
├── docker-compose.yml      # Kafka + Zookeeper setup
├── Makefile                # Run shortcuts
└── README.md               # You're reading it

````

---

## ⚙️ How to Run Locally

### ✅ 1. Start Kafka & Zookeeper

```bash
make start
````

Starts:

- Kafka on `localhost:9092`
- Zookeeper on `localhost:2181`

---

### ✅ 2. Create Topics

```bash
make setup
```

Creates 3 topics with 2 partitions each:

- `orders`, `payments`, `notifications`

---

### ✅ 3. Run Services in Separate Terminals

#### Terminal 1 — Producer

```bash
make producer
```

#### Terminal 2 — Consumer Group 1

```bash
make consumer1
```

#### Terminal 3 — Consumer Group 2

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

| Command          | Purpose                    |
| ---------------- | -------------------------- |
| `make start`     | Starts Kafka and Zookeeper |
| `make setup`     | Creates Kafka topics       |
| `make producer`  | Runs the producer          |
| `make consumer1` | Runs consumer group 1      |
| `make consumer2` | Runs consumer group 2      |

---

## 📦 Dependencies

- Go >= 1.18
- [Sarama](https://github.com/Shopify/sarama) Kafka client
- Docker & Docker Compose

---

## 💡 Real-World Use Cases

- **Event-driven microservices**: Order service, Payment processor, Notification engine
- **Log aggregation**: Collect logs from services and route to different sinks
- **ETL pipelines**: Stream data for transformation and loading

---

## 🔒 Graceful Shutdowns

Each consumer handles `SIGINT` and `SIGTERM`, logs consumption summaries, and exits cleanly. Example:

```
[orders][0] order_created: evt-1721574564890993100
[orders][0] order_updated: evt-1721574564890993115

Consumption Summary:
 orders[0]: 12
 orders[1]: 9
 orders TOTAL: 21
```

---

## 🧪 Extensibility Ideas

- [ ] Add Prometheus metrics & Grafana dashboards
- [ ] Add REST API gateway for producing events
- [ ] Integrate schema registry with Avro/Protobuf
- [ ] Support retry/Dead-letter queues (DLQs)

---

## 🧼 Cleanup

```bash
docker-compose down        # Stop containers
docker-compose down -v     # Stop + remove volumes
```

---

## 📘 License

[MIT License](LICENSE)

---

## 🤝 Credits

Built by engineers for engineers using:

- [Apache Kafka](https://kafka.apache.org/)
- [Sarama](https://github.com/Shopify/sarama)
- Go Modules + Docker

---
