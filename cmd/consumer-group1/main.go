package main

import (
	"context"
	"encoding/json"
	"kafka-demo/internal/kafka"
	"kafka-demo/internal/types"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
)

// Consumer group ID - identifies this consumer group in Kafka
const groupID = "service-group-1"

// Consumer implements sarama.ConsumerGroupHandler
// Used for consuming and tracking messages with stats
type Consumer struct {
	ready chan bool                          // Signals when the consumer is ready
	stats map[string]map[int32]int           // Tracks consumed message count per topic/partition
	mu    sync.Mutex                         // Mutex to protect concurrent map access
}

// Setup is called when a new session begins and partitions are assigned
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	close(c.ready) // Notify main goroutine that the consumer is ready
	log.Printf("%s assigned partitions: %v", groupID, session.Claims())
	return nil
}

// Cleanup is called when a session ends (rebalance or shutdown)
// Used to log stats collected during the session
func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("\nConsumption Summary:")
	for topic, partitions := range c.stats {
		total := 0
		for partition, count := range partitions {
			log.Printf(" %s[%d]: %d", topic, partition, count)
			total += count
		}
		log.Printf(" %s TOTAL: %d", topic, total)
	}
	return nil
}

// ConsumeClaim processes messages from a topic/partition assigned to this consumer
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Initialize stats tracking for this partition
	c.mu.Lock()
	if c.stats[claim.Topic()] == nil {
		c.stats[claim.Topic()] = make(map[int32]int)
	}
	c.stats[claim.Topic()][claim.Partition()] = 0
	c.mu.Unlock()

	// Process messages from the Kafka topic
	for msg := range claim.Messages() {
		var event types.Event

		// Decode the JSON payload into an Event struct
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error decoding %s[%d]: %v", msg.Topic, msg.Partition, err)
			continue
		}

		// Update tracking stats
		c.mu.Lock()
		c.stats[msg.Topic][msg.Partition]++
		c.mu.Unlock()

		// Log the received event
		log.Printf("[%s][%d] %s: %s", msg.Topic, msg.Partition, event.Type, event.ID)

		// Mark message as processed (offset will be committed)
		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	// Create a Sarama consumer configuration with sane defaults
	config := kafka.NewConsumerConfig(groupID)

	// Initialize the consumer handler
	consumer := &Consumer{
		ready: make(chan bool),
		stats: make(map[string]map[int32]int),
	}

	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new Kafka consumer group client
	client, err := sarama.NewConsumerGroup(kafka.GetBrokers(), groupID, config)
	if err != nil {
		log.Panicf("Error creating consumer group: %v", err)
	}

	// Launch consumption loop in a goroutine
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// Consume messages from "orders" and "payments" topics
			if err := client.Consume(ctx, []string{"orders", "payments"}, consumer); err != nil {
				log.Printf("Consumer error: %v", err)
			}

			// Break the loop if the context is cancelled (shutdown triggered)
			if ctx.Err() != nil {
				return
			}

			// Reset readiness channel for rebalances or restarts
			consumer.ready = make(chan bool)
		}
	}()

	// Block until the consumer signals it’s ready
	<-consumer.ready
	log.Printf("%s ready", groupID)

	// Listen for OS termination signals (e.g., CTRL+C or docker stop)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal or context cancel
	select {
	case <-sigterm:
		log.Println("Termination signal received")
	case <-ctx.Done():
		log.Println("Context canceled")
	}

	// Cancel context to shut down consumer
	cancel()

	// Wait for the consumption goroutine to exit
	wg.Wait()

	// Close the consumer group client
	if err := client.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}
}
