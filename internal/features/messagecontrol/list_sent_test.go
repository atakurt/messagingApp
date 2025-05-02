package messagecontrol

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	"github.com/atakurt/messagingApp/internal/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestParseLimit(t *testing.T) {
	tests := []struct {
		input, def, max, expected int
	}{
		{0, 10, 100, 10},
		{-5, 10, 100, 10},
		{5, 10, 100, 5},
		{150, 10, 100, 100},
	}

	for _, tt := range tests {
		got := parseLimit(tt.input, tt.def, tt.max)
		if got != tt.expected {
			t.Errorf("parseLimit(%d, %d, %d) = %d; want %d", tt.input, tt.def, tt.max, got, tt.expected)
		}
	}
}

func TestListSentMessages(t *testing.T) {
	// Create test cases
	tests := []struct {
		name           string
		lastID         int
		limit          int
		setupMock      func(*gomock.Controller) *mocks.MockMessageRepositoryInterface
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:   "Success with messages",
			lastID: 0,
			limit:  10,
			setupMock: func(ctrl *gomock.Controller) *mocks.MockMessageRepositoryInterface {
				mockRepo := mocks.NewMockMessageRepositoryInterface(ctrl)
				messages := []db.Message{
					{
						ID:          1,
						PhoneNumber: "1234567890",
						Content:     "Test message 1",
						Status:      "done",
						MessageID:   "msg-1",
						ProcessedAt: time.Now(),
						SentAt:      time.Now(),
						CreatedAt:   time.Now(),
					},
				}
				mockRepo.EXPECT().GetSentMessages(0, 10).Return(messages, nil)
				return mockRepo
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:   "Success with empty messages",
			lastID: 0,
			limit:  10,
			setupMock: func(ctrl *gomock.Controller) *mocks.MockMessageRepositoryInterface {
				mockRepo := mocks.NewMockMessageRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetSentMessages(0, 10).Return([]db.Message{}, nil)
				return mockRepo
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:   "Repository error",
			lastID: 0,
			limit:  10,
			setupMock: func(ctrl *gomock.Controller) *mocks.MockMessageRepositoryInterface {
				mockRepo := mocks.NewMockMessageRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetSentMessages(0, 10).Return([]db.Message{}, errors.New("database error"))
				return mockRepo
			},
			expectedStatus: fiber.StatusInternalServerError,
		},
		{
			name:   "Custom limit and lastID",
			lastID: 5,
			limit:  20,
			setupMock: func(ctrl *gomock.Controller) *mocks.MockMessageRepositoryInterface {
				mockRepo := mocks.NewMockMessageRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetSentMessages(5, 20).Return([]db.Message{}, nil)
				return mockRepo
			},
			expectedStatus: fiber.StatusOK,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new controller for each test
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create the mock repository
			mockRepo := tt.setupMock(ctrl)

			// Create the service with the mock repository
			service := NewService(mockRepo)

			// Create a new Fiber app
			app := fiber.New()

			// Register the handler
			app.Get("/sent-messages", service.ListSentMessages)

			// Create request with query parameters
			url := "/sent-messages"
			if tt.lastID > 0 || tt.limit > 0 {
				url += "?"
				if tt.lastID > 0 {
					url += "last_id=" + toString(tt.lastID)
				}
				if tt.limit > 0 {
					if tt.lastID > 0 {
						url += "&"
					}
					url += "limit=" + toString(tt.limit)
				}
			}

			// Perform the request
			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req)
			assert.Nil(t, err)

			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// Helper function to convert int to string
func toString(i int) string {
	return fmt.Sprintf("%d", i)
}
