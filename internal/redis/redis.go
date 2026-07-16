package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type CacheService struct {
	Client *redis.Client
}

func NewCacheService(client *redis.Client) *CacheService {
	if client == nil {
		log.Println("⚠️ Redis client is nil, caching disabled")
		return &CacheService{Client: nil}
	}
	return &CacheService{Client: client}
}

func (c *CacheService) Set(key string, data interface{}, ttl time.Duration) error {
	if c.Client == nil {
		return nil // skip cache jika Redis tidak aktif
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Client.Set(ctx, key, jsonData, ttl).Err()
}

func (c *CacheService) Get(key string, dest interface{}) error {
	if c.Client == nil {
		return redis.Nil // skip jika tidak aktif
	}
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *CacheService) Delete(pattern string) error {
	if c.Client == nil {
		return nil
	}
	iter := c.Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.Client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}