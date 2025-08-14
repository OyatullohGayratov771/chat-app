package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.Hash{},
	}
	return &KafkaProducer{writer: w}
}

func (p *KafkaProducer) Publish(ctx context.Context, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}
