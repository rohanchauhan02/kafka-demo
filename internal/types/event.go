package types

import "time"

type Event struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type ConsumerStats struct {
	Topic      string
	Partition  int32
	Count      int
	LastOffset int64
}
