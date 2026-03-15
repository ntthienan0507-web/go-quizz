package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr string, redisURL string) *redis.Client {
	var opts *redis.Options

	if redisURL != "" {
		var err error
		opts, err = redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("failed to parse REDIS_URL: %v", err)
		}
	} else {
		opts = &redis.Options{Addr: addr}
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	log.Println("redis connected")
	return client
}
