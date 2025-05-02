package scheduler

import (
	"context"
	"go.uber.org/zap"
	"sync"
	"time"

	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"

	"github.com/atakurt/messagingApp/internal/features/sendmessages"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
)

type SchedulerInterface interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
	SubscribeToCommands(ctx context.Context)
}

type Scheduler struct {
	ticker         *time.Ticker
	stopChan       chan struct{}
	wg             sync.WaitGroup
	running        bool
	messageService sendmessages.MessageServiceInterface
	redisClient    redisClient.Client
}

func NewScheduler(service sendmessages.MessageServiceInterface, redisClient redisClient.Client) *Scheduler {
	s := &Scheduler{messageService: service, redisClient: redisClient}
	go s.SubscribeToCommands(context.Background())
	return s
}

func (s *Scheduler) SubscribeToCommands(ctx context.Context) {
	pubsub := s.redisClient.Subscribe(ctx, "scheduler:commands")
	defer pubsub.Close()

	// Listen for messages
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			logger.Log.Error("Error receiving message from Redis Pub/Sub", zap.Error(err))
			time.Sleep(time.Second * 5) // Wait before reconnecting
			continue
		}

		switch msg.Payload {
		case "start":
			logger.Log.Info("Received start command via Pub/Sub")
			s.Start(ctx)
		case "stop":
			logger.Log.Info("Received stop command via Pub/Sub")
			s.Stop(ctx)
		default:
			logger.Log.Warn("Unknown scheduler command received", zap.String("payload", msg.Payload))
		}
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	if !config.Cfg.Scheduler.Enabled {
		logger.Log.Warn("Scheduler is disabled by config")
		return
	}

	if s.running {
		logger.Log.Warn("Scheduler already running")
		return
	}

	err := s.redisClient.Set(ctx, "scheduler.enabled", true, 0)
	if err != nil {
		logger.Log.Error("Unable to set scheduler.enabled to true")
	}

	if s.ticker != nil {
		logger.Log.Warn("Scheduler didnt stop yet")
		return
	}
	s.stopChan = make(chan struct{})
	s.ticker = time.NewTicker(config.Cfg.Scheduler.Interval)
	s.running = true
	logger.Log.Info("Scheduler started")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ticker.C:
				if config.Cfg.Scheduler.Enabled {
					logger.Log.Info("Scheduler tick - checking for unsent messages")
					s.messageService.ProcessUnsentMessages(ctx)
				}
			case <-s.stopChan:
				s.ticker.Stop()
				s.ticker = nil
				return
			}
		}
	}()
}

func (s *Scheduler) Stop(ctx context.Context) {
	if !s.running {
		logger.Log.Warn("Scheduler is not running")
		return
	}

	if s.stopChan != nil {
		close(s.stopChan)
		s.wg.Wait()
	}
	s.running = false
	err := s.redisClient.Set(ctx, "scheduler.enabled", false, 0)
	if err != nil {
		logger.Log.Error("Unable to set scheduler.enabled to false")
	}
	logger.Log.Info("Scheduler stopped")
}

func PublishCommand(ctx context.Context, redisClient redisClient.Client, command string) error {
	return redisClient.Publish(ctx, "scheduler:commands", command)
}
