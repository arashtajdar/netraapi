package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// ConnectRedis initializes the Redis connection
func ConnectRedis() {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // fallback for local dev if not set
	}

	opts, err := redis.ParseURL(redisAddr)
	if err != nil {
		// If parsing fails, assume it's just host:port
		opts = &redis.Options{
			Addr: redisAddr,
		}
	}

	RedisClient = redis.NewClient(opts)

	_, err = RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Printf("⚠️  Could not connect to Redis: %v. Caching will be disabled.", err)
		RedisClient = nil
		return
	}

	log.Println("✅ Successfully connected to Redis!")
}

// ClearCachePattern clears all keys matching a specific pattern
func ClearCachePattern(pattern string) {
	if RedisClient == nil {
		return
	}
	iter := RedisClient.Scan(Ctx, 0, pattern, 0).Iterator()
	for iter.Next(Ctx) {
		RedisClient.Del(Ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("Error clearing cache pattern %s: %v", pattern, err)
	}
}
