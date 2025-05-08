package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

//go:generate mockgen -destination=../../mocks/mock_redis.go -package=mocks -mock_names Client=MockRedisClient github.com/atakurt/messagingApp/internal/infrastructure/redis Client
type Client interface {
	Exists(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Subscribe(ctx context.Context, channel string) *PubSub
	Publish(ctx context.Context, channel string, message interface{}) error
	Ping(ctx context.Context) *redis.StatusCmd
	Close(ctx context.Context) error
}

// PubSub represents a Redis Pub/Sub subscription
type PubSub struct {
	pubsub *redis.PubSub
}

// NewPubSub creates a new PubSub with the given redis.PubSub
func NewPubSub(pubsub *redis.PubSub) *PubSub {
	return &PubSub{
		pubsub: pubsub,
	}
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

func (r *RedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}

func (r *RedisClient) Close(ctx context.Context) error {
	return r.client.WithContext(ctx).Close()
}
