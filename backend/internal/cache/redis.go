package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testbox/internal/config"
	"testbox/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

func NewRedisCache(cfg *config.Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✓ Connected to Redis")

	return &RedisCache{
		client: client,
		ctx:    ctx,
		ttl:    15 * time.Minute, // Default TTL
	}, nil
}

// GetTodo implements Lazy Loading: fetch from cache, return nil if not found
func (r *RedisCache) GetTodo(id string) (*models.Todo, error) {
	key := fmt.Sprintf("todo:%s", id)

	data, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		// Cache miss - this is expected in lazy loading
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var todo models.Todo
	if err := json.Unmarshal([]byte(data), &todo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal todo: %w", err)
	}

	log.Printf("✓ Cache HIT: todo:%s", id)
	return &todo, nil
}

// SetTodo implements Write-Through: save to cache when writing to DB
func (r *RedisCache) SetTodo(todo *models.Todo) error {
	key := fmt.Sprintf("todo:%s", todo.ID)

	data, err := json.Marshal(todo)
	if err != nil {
		return fmt.Errorf("failed to marshal todo: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	log.Printf("✓ Cache SET: todo:%s", todo.ID)
	return nil
}

// DeleteTodo removes from cache
func (r *RedisCache) DeleteTodo(id string) error {
	key := fmt.Sprintf("todo:%s", id)

	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	log.Printf("✓ Cache DELETE: todo:%s", id)
	return nil
}

// InvalidateAll clears all todo caches
func (r *RedisCache) InvalidateAll() error {
	iter := r.client.Scan(r.ctx, 0, "todo:*", 0).Iterator()
	for iter.Next(r.ctx) {
		if err := r.client.Del(r.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}

	log.Println("✓ Cache INVALIDATED: all todos")
	return nil
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
