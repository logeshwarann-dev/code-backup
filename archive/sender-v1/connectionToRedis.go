package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
)

func RedisConnection() (*redis.Client, error) {
	podName := os.Getenv("POD_NAME")
	secondaryRedisTemplate := os.Getenv("secondary-redis-template")
	podIndex := strings.Split(podName, "-")[1] // Extract pod index from pod name (e.g., trader-client-0 -> 0)
	secondaryRedisURL := strings.ReplaceAll(secondaryRedisTemplate, "{pod-index}", podIndex)

	rdb := redis.NewClient(&redis.Options{
		Addr: secondaryRedisURL,
		// Addr: "3.94.125.137:6379",
		DB: DB_FOR_MARKET_REPLAY,
	})

	// Test connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis (DB: %d): %v", DB_FOR_MARKET_REPLAY, err)
	}

	fmt.Printf("Connected to Redis (DB: %d)\n", DB_FOR_MARKET_REPLAY)
	return rdb, nil
}
