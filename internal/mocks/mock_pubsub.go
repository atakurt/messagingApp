package mocks

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// MockPubSub is a mock implementation of Redis PubSub
type MockPubSub struct {
	done chan struct{}
}

// NewMockPubSub creates a new mock PubSub
func NewMockPubSub() *MockPubSub {
	return &MockPubSub{
		done: make(chan struct{}),
	}
}

// ReceiveMessage implements the PubSub interface
func (m *MockPubSub) ReceiveMessage(ctx context.Context) (*redis.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-m.done:
		return nil, nil
	}
}

// Close implements the PubSub interface
func (m *MockPubSub) Close() error {
	close(m.done)
	return nil
}
