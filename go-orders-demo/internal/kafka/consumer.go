package kafka

import (
	"context"
	"encoding/json"
	"log"

	"go-orders-demo/internal/db"
	"github.com/segmentio/kafka-go"
)

type Handler func(id string, raw json.RawMessage)

type Consumer struct {
	r *kafka.Reader
	db *db.Store
	h  Handler
}

func NewConsumer(brokers, topic, group string, store db.Repository, h Handler) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		GroupID: group,
		Topic:   topic,
	})
	return &Consumer{r: r, db: store, h: h}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		m, err := c.r.ReadMessage(ctx)
		if err != nil {
			return err
		}
		var tmp map[string]any
		if err := json.Unmarshal(m.Value, &tmp); err != nil {
			log.Printf("skip invalid json: %v", err)
			continue
		}
		id, _ := tmp["order_uid"].(string)
		if id == "" {
			log.Printf("skip without order_uid")
			continue
		}
		if err := c.db.Save(ctx, id, json.RawMessage(m.Value)); err != nil {
			log.Printf("db save: %v", err)
			continue
		}
		if c.h != nil {
			c.h(id, json.RawMessage(m.Value))
		}
	}
}


func (c *Consumer) Close() error { return c.r.Close() }
