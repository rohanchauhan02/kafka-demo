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

const groupID = "service-group-1"

type Consumer struct {
	ready chan bool
	stats map[string]map[int32]int
	mu    sync.Mutex
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	close(c.ready)
	log.Printf("%s assigned partitions: %v", groupID, session.Claims())
	return nil
}

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

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	c.mu.Lock()
	if c.stats[claim.Topic()] == nil {
		c.stats[claim.Topic()] = make(map[int32]int)
	}
	c.stats[claim.Topic()][claim.Partition()] = 0
	c.mu.Unlock()

	for msg := range claim.Messages() {
		var event types.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error decoding %s[%d]: %v", msg.Topic, msg.Partition, err)
			continue
		}

		c.mu.Lock()
		c.stats[msg.Topic][msg.Partition]++
		c.mu.Unlock()

		log.Printf("[%s][%d] %s: %s", msg.Topic, msg.Partition, event.Type, event.ID)
		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	config := kafka.NewConsumerConfig(groupID)
	consumer := &Consumer{
		ready: make(chan bool),
		stats: make(map[string]map[int32]int),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(kafka.GetBrokers(), groupID, config)
	if err != nil {
		log.Panicf("Error creating consumer group: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, []string{"orders", "payments"}, consumer); err != nil {
				log.Printf("Consumer error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	log.Printf("%s ready", groupID)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigterm:
		log.Println("Termination signal received")
	case <-ctx.Done():
		log.Println("Context canceled")
	}
	cancel()
	wg.Wait()

	if err := client.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}
}
