package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go-orders-demo/internal/api"
	"go-orders-demo/internal/cache"
	"go-orders-demo/internal/db"
	kaf "go-orders-demo/internal/kafka"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	httpAddr := getenv("HTTP_ADDR", ":8081")
	brokers  := getenv("KAFKA_BROKERS", "kafka:9092")
	topic    := getenv("KAFKA_TOPIC", "orders")
	group    := getenv("KAFKA_GROUP", "orders-consumer")
	dsn      := getenv("POSTGRES_DSN", "postgres://app:app@localhost:5432/orders?sslmode=disable")
	cacheLim := 1000
	if v := getenv("CACHE_LIMIT", "1000"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cacheLim = n
		}
	}

	store, err := db.New(dsn)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	data, err := os.ReadFile("sample.json")
	if err == nil {
	    var order models.Order
	    if err := json.Unmarshal(data, &order); err == nil {
        	if err := store.SaveOrder(context.Background(), order); err == nil {
            	    log.Printf("✅ Order saved: %s", order.OrderUID)
        	} else {
                    log.Printf("❌ failed to save order: %v", err)
        	}
	    } else {
	        log.Printf("❌ failed to parse json: %v", err)
	    }
	} else {
	    log.Printf("⚠️ sample.json not found (это не ошибка, можно игнорировать)")
	}

	c := cache.New(cacheLim)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	all, err := store.LoadAll(ctx, cacheLim)
	if err != nil {
		log.Printf("warm cache: %v", err)
	} else {
		c.BulkLoad(all)
		log.Printf("warm cache: loaded %d orders", len(all))
	}

	producer := kaf.NewProducer(brokers, topic)
	defer producer.Close()

	srv := api.New(httpAddr, c, store, producer)

	consumer := kaf.NewConsumer(brokers, topic, group, store, func(id string, raw json.RawMessage) {
		c.Set(id, raw)
	})

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("HTTP listen on %s", httpAddr)
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("http: %v", err)
		}
	}()

	go func() {
		log.Printf("Kafka consume on %s topic=%s group=%s", brokers, topic, group)
		if err := consumer.Run(rootCtx); err != nil {
			log.Printf("consumer stopped: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Printf("shutdown...")
	shCtx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	_ = srv.Stop(shCtx)
}
