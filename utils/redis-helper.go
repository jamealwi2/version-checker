package utils

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx context.Context

func init() {
	// Create a new Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: "<REDIS_HOST>:<PORT>",
		DB:   0, // Use default DB
	})
	ctx = context.Background()
}

func HSET(key, field, value string) {
	// HSET key field value
	err := rdb.HSet(ctx, key, field, value).Err()
	if err != nil {
		log.Fatalf("Could not set field: %v", err)
	}
}

func HGETALL(key string) map[string]string {
	// HGETALL key
	values, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		log.Fatalf("Could not get all fields: %v", err)
	}
	return values
}

func HSETAll(key string, values map[string]string) {
	// Convert the map to a slice of interface{} for HSet
	var fieldsAndValues []interface{}
	for field, value := range values {
		fieldsAndValues = append(fieldsAndValues, field, value)
	}

	// HSET key field value
	err := rdb.HSet(ctx, key, fieldsAndValues...).Err()
	if err != nil {
		log.Fatalf("Could not set fields: %v", err)
	}
}
