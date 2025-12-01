package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IsaacDSC/featureflag/pkg/ctxlog"
	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	rdb *redis.Client
}

func NewPublisher(rdb *redis.Client) Publisher {
	return Publisher{
		rdb: rdb,
	}
}

type Payload struct {
	data    any
	attemps int
}

func NewPayload(msg any) Payload {
	return Payload{data: msg}
}

func (p Publisher) Publish(ctx context.Context, channel string, msg Payload) error {
	l := ctxlog.GetLogger(ctx)
	b, err := json.Marshal(msg.data)
	if err != nil {
		l.Error("marshal payload", "error", err)
		return fmt.Errorf("marshal payload: %v", err)
	}

	channelName := fmt.Sprintf("events.fanout.%s", channel)

	if err := p.rdb.Publish(ctx, channelName, b).Err(); err != nil {
		if msg.attemps == 3 {
			l.Error("publish event with error", "channel", channel, "error", err)
			return fmt.Errorf("publish event with error: %v", err)
		}

		msg.attemps++
		log.Printf("error on publisher %s with retry attemps: %d \n", channel, msg.attemps)

		p.Publish(ctx, fmt.Sprintf("events.fanout.%s", channel), msg)
	}

	l.Debug("publish msg", "channel", channel, "msg", msg.data)

	return nil
}
