package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
)

type Subiscriber struct {
	rdb *redis.Client
}

func NewSubscriber(rdb *redis.Client) Subiscriber {
	return Subiscriber{
		rdb: rdb,
	}
}

type Msg []byte

func (m Msg) ToJson(value any) error {
	return json.Unmarshal(m, value)
}

type Handler func(ctx context.Context, msg Msg) error

func (s Subiscriber) Listener(ctx context.Context, channel string, fn Handler) {
	channelName := fmt.Sprintf("events.fanout.%s", channel)
	sub := s.rdb.Subscribe(ctx, channelName)
	defer sub.Close()

	_, err := sub.Receive(ctx)
	if err != nil {
		log.Fatalf("subscribe receive failed: %v", err)
	}

	ch := sub.Channel()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("listening on channel %q", channel)
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				log.Println("subscription closed")
				return
			}

			log.Printf("received: %s", msg.Payload)
			// responder no http

			if err := fn(ctx, []byte(msg.Payload)); err != nil {
				log.Printf("error processing channel %s: %v\n", channel, err)
			}
		case <-sigCh:
			log.Println("shutting down")
			return
		case <-ctx.Done():
			return
		}
	}
}
