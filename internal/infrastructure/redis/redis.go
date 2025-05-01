package redis

import (
	"context"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client
var Ctx = context.Background()

func Init() {
	Rdb = redis.NewClient(&redis.Options{
		Addr: config.Cfg.Redis.Addr,
	})
}
