package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewRetryScheduler(t *testing.T) {
	logger.Log = zap.NewNop()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageRetryServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewRetryScheduler(mockService, mockRedis, getTestConfig())

	// Assertions
	assert.NotNil(t, scheduler)
	assert.Equal(t, mockService, scheduler.service)
	assert.Equal(t, mockRedis, scheduler.redisClient)
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.ticker)
}

func TestRetryScheduler_Start(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageRetryServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)
	mockPubSub := createMockPubSub()

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(mockPubSub).
		AnyTimes()

	mockService.EXPECT().
		ProcessMessageRetries(gomock.Any()).
		AnyTimes()

	scheduler := NewRetryScheduler(mockService, mockRedis, getTestConfig())
	ctx := context.Background()

	// Start scheduler
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	time.Sleep(150 * time.Millisecond)

	// Assertions
	assert.True(t, scheduler.running)
	assert.NotNil(t, scheduler.ticker)
}

func TestRetryScheduler_Stop(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageRetryServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewRetryScheduler(mockService, mockRedis, getTestConfig())
	ctx := context.Background()

	// Start scheduler first
	scheduler.Start(ctx)
	time.Sleep(150 * time.Millisecond)

	// Stop scheduler
	scheduler.Stop(ctx)

	// Assertions
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.ticker)
}

func TestRetryScheduler_Start_WhenAlreadyRunning(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageRetryServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewRetryScheduler(mockService, mockRedis, getTestConfig())
	ctx := context.Background()

	// Start scheduler first time
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)
	assert.True(t, scheduler.running)
	time.Sleep(150 * time.Millisecond)

	// Try to start again
	scheduler.Start(ctx)
	assert.True(t, scheduler.running) // Should still be running
}

func TestRetryScheduler_Stop_WhenNotRunning(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageRetryServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewRetryScheduler(mockService, mockRedis, getTestConfig())
	ctx := context.Background()

	// Try to stop when not running
	scheduler.Stop(ctx)
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.ticker)
}

func TestRetryScheduler_SubscribeToCommands(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageRetryServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockPubSub := NewMockPubSub()
	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(mockPubSub).
		AnyTimes()

	mockRedis.EXPECT().
		Publish(gomock.Any(), "scheduler:commands", "start").
		Return(nil).
		Times(1)

	scheduler := NewRetryScheduler(mockService, mockRedis, getTestConfig())
	scheduler.Start(ctx)

	// Test PublishCommand
	err := PublishCommand(ctx, mockRedis, "start")
	assert.NoError(t, err)

	// Wait for context to timeout
	<-ctx.Done()

	// Clean up
	scheduler.Stop(ctx)
}

func TestRetryScheduler_SubscribeToCommands_UnknownCommand(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageRetryServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	mockRedis.EXPECT().
		Publish(gomock.Any(), "scheduler:commands", "unknown").
		Return(nil).
		Times(1)

	scheduler := NewRetryScheduler(mockService, mockRedis, getTestConfig())
	scheduler.Start(ctx)

	// Test PublishCommand with unknown command
	err := PublishCommand(ctx, mockRedis, "unknown")
	assert.NoError(t, err)

	// Wait for context to timeout
	<-ctx.Done()

	// Clean up
	scheduler.Stop(ctx)
}

func getTestConfig() config.Config {
	return config.Config{
		Scheduler: struct {
			Enabled            bool
			Interval           time.Duration
			BatchSize          int
			MaxConcurrent      int
			MaxRetryConcurrent int
		}{
			Enabled:  true,
			Interval: time.Second,
		},
	}
}
