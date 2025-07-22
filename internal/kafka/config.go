package kafka

import "github.com/IBM/sarama"

// NewProducerConfig returns a configured Sarama config for Kafka producers.
// It enables acknowledgments, hashing-based partitioning, and delivery confirmations.
func NewProducerConfig() *sarama.Config {
	config := sarama.NewConfig()

	// Ensures delivery reports for successful messages are returned on the Successes channel.
	config.Producer.Return.Successes = true

	// Uses a hash partitioner to consistently route messages with the same key to the same partition.
	config.Producer.Partitioner = sarama.NewHashPartitioner

	// Waits for acknowledgment from all in-sync replicas before considering a message as "produced".
	config.Producer.RequiredAcks = sarama.WaitForAll

	return config
}

// NewConsumerConfig returns a configured Sarama config for Kafka consumer groups.
// It sets group rebalancing strategy, offset handling, and error return behavior.
func NewConsumerConfig(groupID string) *sarama.Config {
	config := sarama.NewConfig()

	// Ensures consumer errors are returned on the Errors channel.
	config.Consumer.Return.Errors = true

	// Starts consuming from the oldest available offset if no committed offset is found.
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Uses range-based partition assignment within the group.
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRange(),
	}

	// Sets the client ID for logging and metrics tracking.
	config.ClientID = groupID

	return config
}

// GetBrokers returns the list of Kafka broker addresses.
func GetBrokers() []string {
	return []string{"localhost:9092"}
}

// GetTopics returns the list of Kafka topics to be used across the system.
func GetTopics() []string {
	return []string{"orders", "payments", "notifications"}
}
