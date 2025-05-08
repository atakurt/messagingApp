package commandlistener

import (
	"context"
	"errors"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"go.uber.org/zap"
	"time"
)

//go:generate mockgen -destination=../../../mocks/mock_controllable.go -package=mocks github.com/atakurt/messagingApp/internal/features/messagecontrol/commandlistener Controllable
type Controllable interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}

type CommandListener struct {
	redisClient *redis.RedisClient
	targets     []Controllable
}

func NewCommandListener(redisClient *redis.RedisClient, targets ...Controllable) *CommandListener {
	return &CommandListener{
		redisClient: redisClient,
		targets:     targets,
	}
}

func (d *CommandListener) Listen(ctx context.Context) {
	pubsub := d.redisClient.Subscribe(ctx, "scheduler:commands")
	defer pubsub.Close()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Command dispatcher exiting due to context cancellation")
			return
		default:
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					logger.Log.Info("Context cancelled, stopping dispatcher")
					return
				}
				logger.Log.Error("Error receiving Pub/Sub message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			logger.Log.Info("Received scheduler command", zap.String("payload", msg.Payload))

			switch msg.Payload {
			case "start":
				for _, t := range d.targets {
					t.Start(ctx)
				}
			case "stop":
				for _, t := range d.targets {
					t.Stop(ctx)
				}
			default:
				logger.Log.Warn("Unknown command", zap.String("command", msg.Payload))
			}
		}
	}
}
