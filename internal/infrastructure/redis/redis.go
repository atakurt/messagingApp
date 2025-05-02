package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client interface {
	Exists(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Subscribe(ctx context.Context, channel string) *PubSub
	Publish(ctx context.Context, channel string, message interface{}) error
}

// PubSub represents a Redis Pub/Sub subscription
type PubSub struct {
	pubsub *redis.PubSub
}

// ReceiveMessage receives a message from the subscription
func (p *PubSub) ReceiveMessage(ctx context.Context) (*redis.Message, error) {
	return p.pubsub.ReceiveMessage(ctx)
}

// Close closes the subscription
func (p *PubSub) Close() error {
	return p.pubsub.Close()
}

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewClient(ctx context.Context, client *redis.Client) *RedisClient {
	return &RedisClient{
		ctx:    ctx,
		client: client,
	}
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result == 1, err
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *RedisClient) Subscribe(ctx context.Context, channel string) *PubSub {
	pubsub := r.client.Subscribe(ctx, channel)
	return &PubSub{
		pubsub: pubsub,
	}
}

func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}
