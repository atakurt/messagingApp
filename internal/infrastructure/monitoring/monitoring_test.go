package monitoring

import (
	"errors"
	"testing"

	"github.com/atakurt/messagingApp/internal/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestNewMonitoringService(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDBInterface(ctrl)
	mockRedisClient := mocks.NewMockRedisClient(ctrl)

	// when
	service := NewMonitoringService(mockDB, mockRedisClient)

	// then
	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockRedisClient, service.redis)
}

func TestMonitoringService_Liveness(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDBInterface(ctrl)
	mockRedisClient := mocks.NewMockRedisClient(ctrl)
	service := NewMonitoringService(mockDB, mockRedisClient)

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	// when
	err := service.Liveness(ctx)

	// then
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, ctx.Response().StatusCode())
}

func TestMonitoringService_Readiness_DBConnectionError(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)

	mockRedisClient := mocks.NewMockRedisClient(ctrl)
	mockDB := mocks.NewMockDBInterface(ctrl)
	mockDB.EXPECT().GetSQLDB().Return(nil, errors.New("DB connection error"))

	service := NewMonitoringService(mockDB, mockRedisClient)

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	// when
	err := service.Readiness(ctx)

	// then
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, ctx.Response().StatusCode())
}
