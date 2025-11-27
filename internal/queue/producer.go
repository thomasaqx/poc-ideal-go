package queue

import (
	"context"
	"github.com/segmentio/kafka-go"
	"time"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokerUrl string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerUrl),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) Publish(msg string) error {
	// Short timeout to avoid blocking the API
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return p.writer.WriteMessages(ctx,
		kafka.Message{
			Value: []byte(msg),
		},
	)
}

func (p *KafkaProducer) Close() {
	p.writer.Close()
}