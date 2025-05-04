package monitoring

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/atakurt/messagingApp/internal/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

type MockDB struct {
	sqlDB   *sql.DB
	dbErr   error
	pingErr error
}

func (m *MockDB) GetSQLDB() (*sql.DB, error) {
	if m.dbErr != nil {
		return nil, m.dbErr
	}
	return nil, nil
}

func (m *MockDB) GetDB() *gorm.DB {
	return nil
}

func (m *MockDB) Begin() *gorm.DB {
	return nil
}

func TestNewMonitoringService(t *testing.T) {
	// given
	mockDB := &MockDB{}
	mockRedisClient := mocks.NewMockRedisClient(gomock.NewController(t))

	// when
	service := NewMonitoringService(mockDB, mockRedisClient)

	// then
	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockRedisClient, service.redis)
}

func TestMonitoringService_Liveness(t *testing.T) {
	// given
	mockDB := &MockDB{}
	mockRedisClient := mocks.NewMockRedisClient(gomock.NewController(t))
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
	defer ctrl.Finish()

	mockRedisClient := mocks.NewMockRedisClient(ctrl)
	mockDB := &MockDB{
		dbErr: errors.New("DB connection error"),
	}

	service := NewMonitoringService(mockDB, mockRedisClient)

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	// when
	err := service.Readiness(ctx)

	// then
	assert.NoError(t, err) // The method returns no error, but sets status code
	assert.Equal(t, fiber.StatusServiceUnavailable, ctx.Response().StatusCode())
}
