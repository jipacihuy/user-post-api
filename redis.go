package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

// InitRedis koneksi ke Redis
func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		//redisURL = "redis://localhost:6379"
		redisURL = "redis://127.0.0.1:6379"

	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Failed to parse REDIS_URL:", err)
	}

	rdb = redis.NewClient(opt)

	// Test koneksi
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("✅ Redis connected")
}

// SetCache menyimpan data ke Redis dengan TTL
func SetCache(key string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ Marshal error: %v", err)
		return err
	}
	err = rdb.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		log.Printf("❌ Redis Set error: %v", err)
	} else {
		log.Printf("✅ Cache saved: %s", key)
	}
	return err
}

// GetCache mengambil data dari Redis
func GetCache(key string, dest interface{}) error {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// DeleteCache menghapus cache dengan pattern tertentu
func DeleteCache(pattern string) error {
	iter := rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		err := rdb.Del(ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	return iter.Err()
}
