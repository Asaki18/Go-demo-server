package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kafka.Writer
}

func NewProducer(brokers, topic string) *Producer {
	return &Producer{w: &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
	}}
}

func (p *Producer) Produce(ctx context.Context, payload []byte) error {
	return p.w.WriteMessages(ctx, kafka.Message{Value: payload})
}

func (p *Producer) Close() error { return p.w.Close() }