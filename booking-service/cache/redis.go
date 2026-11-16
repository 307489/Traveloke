package cache

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var ctx = context.Background()

func Connect() {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "localhost:6379"
	}
	client = redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(ctx).Err(); err != nil {
		log.Println("Redis unavailable, cache disabled:", err)
		client = nil
		return
	}
	log.Println("Connected to Redis")
}


func Get(key string, dest any) bool {
	if client == nil {
		return false
	}
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return json.Unmarshal([]byte(val), dest) == nil
}

func Set(key string, value any, ttl time.Duration) {
	if client == nil {
		return
	}
	b, err := json.Marshal(value)
	if err != nil {
		return
	}
	client.Set(ctx, key, b, ttl)
}


func InvalidatePrefix(prefix string) {
	if client == nil {
		return
	}
	iter := client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		client.Del(ctx, iter.Val())
	}
}