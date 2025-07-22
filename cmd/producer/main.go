package main

import (
	"encoding/json"
	"fmt"
	"kafka-demo/internal/kafka"
	"kafka-demo/internal/types"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

func main() {
	// Initialize Kafka producer configuration using custom helper
	config := kafka.NewProducerConfig()

	// Create an asynchronous Kafka producer
	producer, err := sarama.NewAsyncProducer(kafka.GetBrokers(), config)
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer producer.Close()

	// Channel to listen for OS interrupt signals (e.g., Ctrl+C)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// WaitGroup to ensure logging goroutines complete before exit
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to handle success acknowledgments from Kafka
	go func() {
		defer wg.Done()
		for msg := range producer.Successes() {
			log.Printf("Produced to %s[%d]", msg.Topic, msg.Partition)
		}
	}()

	// Goroutine to handle errors from Kafka broker
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Printf("Producer error: %v", err)
		}
	}()

	// Define possible event types for each topic
	eventTypes := map[string][]string{
		"orders":        {"order_created", "order_updated", "order_canceled"},
		"payments":      {"payment_received", "payment_failed", "payment_refunded"},
		"notifications": {"email_sent", "sms_sent", "push_notification"},
	}

ProducerLoop:
	for {
		// Randomly select a topic and corresponding event type
		topic := kafka.GetTopics()[rand.Intn(3)]
		eventType := eventTypes[topic][rand.Intn(3)]
		eventID := fmt.Sprintf("evt-%d", time.Now().UnixNano())

		// Construct an event struct
		event := types.Event{
			Type:      eventType,
			ID:        eventID,
			Data:      fmt.Sprintf("data-%d", rand.Intn(1000)),
			Timestamp: time.Now(),
		}

		// Serialize the event into JSON format
		jsonEvent, _ := json.Marshal(event)

		// Create a Kafka message with topic, key, and serialized value
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(eventID),       // Used for partitioning
			Value: sarama.ByteEncoder(jsonEvent),      // JSON payload
		}

		// Send the message or gracefully shut down on interrupt signal
		select {
		case producer.Input() <- msg:
			log.Printf("Sent %s to %s", event.Type, topic)
		case <-signals:
			producer.AsyncClose() // Closes the producer gracefully
			break ProducerLoop
		}

		// Sleep for 500ms to 1.5s before sending the next message
		time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
	}

	// Wait for all logs from success/error goroutines to finish
	wg.Wait()
	log.Println("Producer shutdown complete")
}
