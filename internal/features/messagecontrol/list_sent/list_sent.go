package list_sent

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ListSentServiceInterface interface {
	ListSentMessages(c *fiber.Ctx) error
}

type MessageRepositoryInterface interface {
	GetSentMessages(lastID, limit int) ([]db.Message, error)
}

type ListSentService struct {
	repository MessageRepositoryInterface
}

func NewService(repository MessageRepositoryInterface) *ListSentService {
	return &ListSentService{
		repository: repository,
	}
}

// MessageResponse represents a single message in the response
// @Description Message data structure
type MessageResponse struct {
	ID          uint      `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	MessageID   string    `json:"message_id"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
	SentAt      time.Time `json:"sent_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListResponse represents the paginated response structure
// @Description Paginated list of messages
type ListResponse struct {
	LastID int               `json:"last_id"`
	Limit  int               `json:"limit"`
	Data   []MessageResponse `json:"data"`
}

// ListSentMessages godoc
// @Summary      List sent messages
// @Description  Retrieves all messages that have been marked as sent using offset-based pagination
// @Tags         Messages
// @Produce      json
// @Param        last_id  query     int  false  "Only return messages with ID > last_id"
// @Param        limit    query     int  false  "Maximum number of messages to return (max 100)"
// @Success      200  {object}  ListResponse
// @Failure      500  {object}  map[string]string
// @Router       /sent-messages [get]
func (l *ListSentService) ListSentMessages(c *fiber.Ctx) error {
	lastID := c.QueryInt("last_id", 0)
	limit := parseLimit(c.QueryInt("limit", 0), 10, 100)

	messages, err := l.repository.GetSentMessages(lastID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve messages",
		})
	}

	// Convert db.Message to MessageResponse
	var responseMessages []MessageResponse
	for _, msg := range messages {
		responseMessages = append(responseMessages, MessageResponse{
			ID:          msg.ID,
			PhoneNumber: msg.PhoneNumber,
			Content:     msg.Content,
			Status:      string(msg.Status),
			MessageID:   msg.MessageID,
			ProcessedAt: msg.ProcessedAt,
			SentAt:      msg.SentAt,
			CreatedAt:   msg.CreatedAt,
		})
	}

	return c.JSON(ListResponse{
		LastID: lastID,
		Limit:  limit,
		Data:   responseMessages,
	})
}

func parseLimit(input, defaultLimit, maxLimit int) int {
	if input <= 0 {
		return defaultLimit
	}
	if input > maxLimit {
		return maxLimit
	}
	return input
}
