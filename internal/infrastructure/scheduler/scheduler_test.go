package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/atakurt/messagingApp/internal/mocks"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type MockPubSub struct {
	done chan struct{}
}

func NewMockPubSub() *redisClient.PubSub {
	return redisClient.NewPubSub(&goRedis.PubSub{})
}

func (m *MockPubSub) ReceiveMessage(ctx context.Context) (*goRedis.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-m.done:
		return nil, nil
	}
}

func (m *MockPubSub) Close() error {
	close(m.done)
	return nil
}

func createMockPubSub() *redisClient.PubSub {
	return NewMockPubSub()
}

func TestNewScheduler(t *testing.T) {
	logger.Log = zap.NewNop()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)

	// Assertions
	assert.NotNil(t, scheduler)
	assert.Equal(t, mockService, scheduler.messageService)
	assert.Equal(t, mockRedis, scheduler.redisClient)
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.ticker)
	assert.Nil(t, scheduler.stopChan)

	// Clean up
	scheduler.Stop(context.Background())
}

func TestScheduler_Start(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)
	mockPubSub := createMockPubSub()

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(mockPubSub).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)
	ctx := context.Background()

	// Set config to enable scheduler
	config.Cfg.Scheduler.Enabled = true
	config.Cfg.Scheduler.Interval = time.Second

	// Start scheduler
	scheduler.Start(ctx)

	// Give some time for the scheduler to start
	time.Sleep(100 * time.Millisecond)

	// Assertions
	assert.True(t, scheduler.running)
	assert.NotNil(t, scheduler.ticker)
	assert.NotNil(t, scheduler.stopChan)

	// Clean up
	scheduler.Stop(ctx)
}

func TestScheduler_Stop(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)
	ctx := context.Background()
	scheduler.Start(ctx)

	// Stop scheduler
	scheduler.Stop(ctx)

	// Assertions
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.ticker)
	assert.Nil(t, scheduler.stopChan)
}

func TestScheduler_SubscribeToCommands(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	// Create a context with timeout
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

	scheduler := NewScheduler(mockService, mockRedis)

	// Test PublishCommand
	err := PublishCommand(ctx, mockRedis, "start")
	assert.NoError(t, err)

	// Wait for context to timeout
	<-ctx.Done()

	// Clean up
	scheduler.Stop(ctx)
}

func TestScheduler_Start_WhenDisabled(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a new config instance for this test
	testConfig := &config.Config{
		Scheduler: struct {
			Enabled            bool
			Interval           time.Duration
			BatchSize          int
			MaxConcurrent      int
			MaxRetryConcurrent int
		}{
			Enabled: false,
		},
	}

	// Save original config and restore after test
	originalConfig := config.Cfg
	defer func() { config.Cfg = originalConfig }()

	// Set test config
	config.Cfg = *testConfig

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)
	ctx := context.Background()

	scheduler.Start(ctx)

	// Assertions
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.ticker)
	assert.Nil(t, scheduler.stopChan)

	// Clean up
	scheduler.Stop(ctx)
}

func TestScheduler_Start_WhenAlreadyRunning(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)
	ctx := context.Background()

	config.Cfg.Scheduler.Enabled = true
	config.Cfg.Scheduler.Interval = time.Second

	// Start scheduler first time
	scheduler.Start(ctx)
	assert.True(t, scheduler.running)

	// Try to start again
	scheduler.Start(ctx)
	assert.True(t, scheduler.running) // Should still be running

	// Clean up
	scheduler.Stop(ctx)
}

func TestScheduler_Stop_WhenNotRunning(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)
	ctx := context.Background()

	// Try to stop when not running
	scheduler.Stop(ctx)
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.stopChan)
	assert.Nil(t, scheduler.ticker)
}

func TestScheduler_SubscribeToCommands_UnknownCommand(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
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

	scheduler := NewScheduler(mockService, mockRedis)

	// Test PublishCommand with unknown command
	err := PublishCommand(ctx, mockRedis, "unknown")
	assert.NoError(t, err)

	// Wait for context to timeout
	<-ctx.Done()

	// Clean up
	scheduler.Stop(ctx)
}

func TestScheduler_SubscribeToCommands_ErrorHandling(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(redisClient.NewPubSub(&goRedis.PubSub{})).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)

	// Wait for context to timeout
	<-ctx.Done()

	// Clean up
	scheduler.Stop(ctx)
}

func TestScheduler_MultipleStartStopCycles(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)
	ctx := context.Background()

	config.Cfg.Scheduler.Enabled = true
	config.Cfg.Scheduler.Interval = time.Second

	// First start-stop cycle
	scheduler.Start(ctx)
	assert.True(t, scheduler.running)
	scheduler.Stop(ctx)
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.stopChan)
	assert.Nil(t, scheduler.ticker)

	// Second start-stop cycle
	scheduler.Start(ctx)
	assert.True(t, scheduler.running)
	scheduler.Stop(ctx)
	assert.False(t, scheduler.running)
	assert.Nil(t, scheduler.stopChan)
	assert.Nil(t, scheduler.ticker)
}

func TestScheduler_Start_WithTickerNotNil(t *testing.T) {
	logger.Log = zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockMessageServiceInterface(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	mockRedis.EXPECT().
		Subscribe(gomock.Any(), "scheduler:commands").
		Return(createMockPubSub()).
		AnyTimes()

	scheduler := NewScheduler(mockService, mockRedis)
	ctx := context.Background()

	// Enable scheduler
	config.Cfg.Scheduler.Enabled = true
	config.Cfg.Scheduler.Interval = time.Second

	scheduler.ticker = time.NewTicker(time.Second)

	scheduler.Start(ctx)
	assert.False(t, scheduler.running) // Should not start because ticker is not nil

	// Clean up
	scheduler.Stop(ctx)
}
