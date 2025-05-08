package retry

import (
	"context"
	"time"

	"github.com/atakurt/messagingApp/internal/features/messageretry"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/infrastructure/redis"
	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
)

type RetrySchedulerInterface interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}

type RetryScheduler struct {
	service     messageretry.MessageRetryServiceInterface
	redisClient redis.Client
	ticker      *time.Ticker
	running     bool
	cfg         config.Config
}

func NewRetryScheduler(service messageretry.MessageRetryServiceInterface, redisClient redis.Client, cfg config.Config) *RetryScheduler {
	return &RetryScheduler{
		service:     service,
		redisClient: redisClient,
		running:     false,
		cfg:         cfg,
	}
}

func (s *RetryScheduler) Start(ctx context.Context) {
	if s.running {
		logger.Log.Warn("Retry scheduler already running")
		return
	}

	// Only start the processing goroutine if scheduler is enabled
	if s.cfg.Scheduler.Enabled {
		s.startProcessing(ctx)
	} else {
		logger.Log.Info("Retry scheduler started in disabled state, waiting for enable command")
	}
}

func (s *RetryScheduler) startProcessing(ctx context.Context) {
	s.running = true
	s.ticker = time.NewTicker(s.cfg.Scheduler.Interval)

	// Start the processing goroutine
	go func() {
		defer func() {
			if s.ticker != nil {
				s.ticker.Stop()
			}
		}()

		for {
			select {
			case <-s.ticker.C:
				if s.running {
					s.service.ProcessMessageRetries(ctx)
				}
			case <-ctx.Done():
				logger.Log.Info("Retry scheduler stopped due to context cancellation")
				return
			}
		}
	}()

	logger.Log.Info("Retry scheduler started with processing enabled")
}

func (s *RetryScheduler) Stop(ctx context.Context) {
	if !s.running {
		logger.Log.Warn("Retry scheduler is not running")
		return
	}

	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}
	logger.Log.Info("Retry scheduler stopped")
}

func publishCommand(ctx context.Context, redisClient redisClient.Client, command string) error {
	return redisClient.Publish(ctx, "scheduler:commands", command)
}
