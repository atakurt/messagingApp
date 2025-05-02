package sendmessages

import (
	"github.com/atakurt/messagingApp/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMessageRepositoryInterface(ctrl)
	mockHttp := mocks.NewMockClient(ctrl)
	mockRedis := mocks.NewMockRedisClient(ctrl)

	service := NewService(mockRepo, mockHttp, mockRedis)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repository)
	assert.Equal(t, mockHttp, service.httpClient)
	assert.Equal(t, mockRedis, service.redisClient)
}
