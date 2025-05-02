package scheduler

import (
	"context"
	"time"

	"github.com/atakurt/messagingApp/internal/features/messageretry"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"go.uber.org/zap"
)

type RetryScheduler struct {
	service     messageretry.MessageRetryServiceInterface
	redisClient redis.Client
	ticker      *time.Ticker
	running     bool
}

func NewRetryScheduler(service messageretry.MessageRetryServiceInterface, redisClient redis.Client) *RetryScheduler {
	return &RetryScheduler{
		service:     service,
		redisClient: redisClient,
		running:     false,
	}
}

func (s *RetryScheduler) Start(ctx context.Context) {
	s.running = true
	s.ticker = time.NewTicker(config.Cfg.Scheduler.Interval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				if s.running {
					s.service.ProcessMessageRetries(ctx)
				}
			case <-ctx.Done():
				logger.Log.Info("Retry scheduler stopped due to context cancellation")
				s.ticker.Stop()
				return
			}
		}
	}()

	// Subscribe to control commands
	go s.subscribeToCommands(ctx)

	logger.Log.Info("Retry scheduler started")
}

func (s *RetryScheduler) subscribeToCommands(ctx context.Context) {
	pubsub := s.redisClient.Subscribe(ctx, "scheduler:commands")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			logger.Log.Error("Failed to receive message from Redis", zap.Error(err))
			continue
		}

		switch msg.Payload {
		case "start":
			s.running = true
			logger.Log.Info("Retry scheduler started via Redis command")
		case "stop":
			s.running = false
			logger.Log.Info("Retry scheduler stopped via Redis command")
		default:
			logger.Log.Warn("Unknown command received", zap.String("command", msg.Payload))
		}
	}
}
