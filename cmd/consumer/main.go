package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yohnnn/booking_service/internal/event"
)

func main() {
	brokers := []string{"localhost:29092"}
	topic := "bookings.created"
	groupID := "notification-service"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Notification consumer started")

	go func() {
		for {
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("fetch error: %v\n", err)
				continue
			}

			if err := handleMessage(ctx, msg); err != nil {
				log.Printf("processing failed (will retry): %v\n", err)
				continue
			}

			if err := reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("commit failed: %v\n", err)
			}
		}
	}()

	<-sig
	log.Println("Shutting down consumer...")
}

func handleMessage(ctx context.Context, msg kafka.Message) error {
	var evt event.BookingCreatedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return err
	}

	log.Printf(
		"Notify user=%s booking=%s concert=%s seat=%d",
		evt.UserID,
		evt.BookingID,
		evt.ConcertID,
		evt.Seat,
	)

	time.Sleep(200 * time.Millisecond)

	return nil
}
