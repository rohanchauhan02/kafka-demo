package kafka

import "github.com/IBM/sarama"

func NewProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	return config
}

func NewConsumerConfig(groupID string) *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRange(),
	}
	config.ClientID = groupID
	return config
}

func GetBrokers() []string {
	return []string{"localhost:9092"}
}

func GetTopics() []string {
	return []string{"orders", "payments", "notifications"}
}
