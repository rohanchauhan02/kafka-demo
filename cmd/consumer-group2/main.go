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

// Consumer group ID used to identify this service
const groupID = "service-group-2"

// Consumer implements the sarama.ConsumerGroupHandler interface
type Consumer struct {
	ready chan bool                          // Signals readiness to consume
	stats map[string]map[int32]int           // Tracks message count per topic-partition
	mu    sync.Mutex                         // Protects concurrent access to stats
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	close(c.ready) // Signal that consumer is ready
	log.Printf("%s assigned partitions: %v", groupID, session.Claims())
	return nil
}

// Cleanup is run at the end of a session, used here to log consumption summary
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

// ConsumeClaim processes messages from the assigned topic/partition
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Initialize the partition counter
	c.mu.Lock()
	if c.stats[claim.Topic()] == nil {
		c.stats[claim.Topic()] = make(map[int32]int)
	}
	c.stats[claim.Topic()][claim.Partition()] = 0
	c.mu.Unlock()

	// Process messages in a streaming loop
	for msg := range claim.Messages() {
		var event types.Event

		// Attempt to decode the message payload into an Event struct
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error decoding %s[%d]: %v", msg.Topic, msg.Partition, err)
			continue
		}

		// Update stats
		c.mu.Lock()
		c.stats[msg.Topic][msg.Partition]++
		c.mu.Unlock()

		// Log the message content
		log.Printf("[%s][%d] %s: %s", msg.Topic, msg.Partition, event.Type, event.ID)

		// Mark message as processed (committed offset)
		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	// Create Sarama consumer configuration
	config := kafka.NewConsumerConfig(groupID)

	// Initialize our consumer handler with stats tracking
	consumer := &Consumer{
		ready: make(chan bool),
		stats: make(map[string]map[int32]int),
	}

	// Create a cancellable context to support graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Create the Kafka consumer group client
	client, err := sarama.NewConsumerGroup(kafka.GetBrokers(), groupID, config)
	if err != nil {
		log.Panicf("Error creating consumer group: %v", err)
	}

	// Launch the consumption loop in a goroutine
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// Start consuming messages from the "payments" and "notifications" topics
			if err := client.Consume(ctx, []string{"payments", "notifications"}, consumer); err != nil {
				log.Printf("Consumer error: %v", err)
			}

			// Exit loop if context is canceled
			if ctx.Err() != nil {
				return
			}

			// Reset readiness channel for next session
			consumer.ready = make(chan bool)
		}
	}()

	// Wait until consumer is ready
	<-consumer.ready
	log.Printf("%s ready", groupID)

	// Listen for system termination signals for graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigterm:
		log.Println("Termination signal received")
	case <-ctx.Done():
		log.Println("Context canceled")
	}

	// Stop the consumer context
	cancel()

	// Wait for the consumption goroutine to finish
	wg.Wait()

	// Close the consumer group client
	if err := client.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}
}
