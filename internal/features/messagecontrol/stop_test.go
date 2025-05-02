package messagecontrol

import (
	"errors"
	"testing"

	"github.com/atakurt/messagingApp/internal/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestStopHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*gomock.Controller) *mocks.MockRedisClient
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mockRedis := mocks.NewMockRedisClient(ctrl)
				mockRedis.EXPECT().Publish(gomock.Any(), "scheduler:commands", "stop").Return(nil)
				return mockRedis
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `{"message":"Stop command sent to all scheduler instances"}`,
		},
		{
			name: "Redis error",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockRedisClient {
				mockRedis := mocks.NewMockRedisClient(ctrl)
				mockRedis.EXPECT().Publish(gomock.Any(), "scheduler:commands", "stop").Return(errors.New("redis connection failed"))
				return mockRedis
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"scheduler operation 'stop' failed: redis connection failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRedis := tt.setupMock(ctrl)

			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			StopHandler(ctx, mockRedis)

			assert.Equal(t, tt.expectedStatus, ctx.Response().StatusCode())

			responseBody := string(ctx.Response().Body())
			assert.JSONEq(t, tt.expectedBody, responseBody)
		})
	}
}
