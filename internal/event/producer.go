package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type EventProducer interface {
	SendBookingCreated(ctx context.Context, event BookingCreatedEvent) error
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
		},
	}
}

func (p *KafkaProducer) SendBookingCreated(ctx context.Context, event BookingCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.BookingID),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
