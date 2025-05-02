package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	goRedis "github.com/go-redis/redis/v8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start Redis container
	ctx := context.Background()
	redisContainer, err := startRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	redisClient := NewClient(ctx, goRedis.NewClient(&goRedis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port.Int()),
	}))

	// Test Publish
	t.Run("Publish", func(t *testing.T) {
		channel := "test-channel"
		message := "test-message"

		// Publish message
		err := redisClient.Publish(ctx, channel, message)
		assert.NoError(t, err)
	})

	// Test Set and Exists
	t.Run("Set and Exists", func(t *testing.T) {
		key := "test-key"
		value := "test-value"

		// Check key doesn't exist initially
		exists, err := redisClient.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)

		// Set value
		err = redisClient.Set(ctx, key, value, 10*time.Second)
		assert.NoError(t, err)

		// Check key exists after setting
		exists, err = redisClient.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	// Test SetNX
	t.Run("SetNX", func(t *testing.T) {
		key := "test-nx-key"
		value := "test-value"

		// First SetNX should succeed
		success, err := redisClient.SetNX(ctx, key, value, 10*time.Second)
		assert.NoError(t, err)
		assert.True(t, success)

		// Second SetNX with same key should fail
		success, err = redisClient.SetNX(ctx, key, "another-value", 10*time.Second)
		assert.NoError(t, err)
		assert.False(t, success)
	})

	// Test Subscribe
	t.Run("Subscribe", func(t *testing.T) {
		channel := "test-subscribe-channel"
		message := "test-subscribe-message"

		// Create a channel to receive messages
		messagesCh := make(chan string, 1)

		// Subscribe to channel in a goroutine
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pubsub := redisClient.Subscribe(ctx, channel)
		defer pubsub.Close()

		go func() {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err == nil {
				messagesCh <- msg.Payload
			}
		}()

		// Wait for subscription to be established
		time.Sleep(100 * time.Millisecond)

		// Publish message
		err := redisClient.Publish(ctx, channel, message)
		assert.NoError(t, err)

		// Wait for message to be received
		select {
		case receivedMsg := <-messagesCh:
			assert.Equal(t, message, receivedMsg)
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for message")
		}
	})

	// Test PubSub Close
	t.Run("PubSub Close", func(t *testing.T) {
		channel := "test-pubsub-close"

		// Create subscription
		pubsub := redisClient.Subscribe(ctx, channel)

		// Close subscription
		err := pubsub.Close()
		assert.NoError(t, err)

		// Attempting to receive after close should fail
		_, err = pubsub.ReceiveMessage(ctx)
		assert.Error(t, err)
	})
}

// Helper function to start a Redis container
func startRedisContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7.4.3",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
