package messagecontrol

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	"github.com/atakurt/messagingApp/internal/infrastructure/repository"
	"github.com/gofiber/fiber/v2"
)

type ListSentServiceInterface interface {
	ListSentMessages(c *fiber.Ctx) error
}

type ListSentService struct {
	repository *repository.MessageRepository
}

func NewService(repository *repository.MessageRepository) *ListSentService {
	return &ListSentService{
		repository: repository,
	}
}

// ListSentMessages godoc
// @Summary      List sent messages
// @Description  Retrieves all messages that have been marked as sent using offset-based pagination
// @Tags         Messages
// @Produce      json
// @Param        last_id  query     int  false  "Only return messages with ID > last_id"
// @Param        limit    query     int  false  "Maximum number of messages to return (max 100)"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /sent-messages [get]
func (l *ListSentService) ListSentMessages(c *fiber.Ctx) error {
	lastID := c.QueryInt("last_id", 0)
	limit := parseLimit(c.QueryInt("limit", 0), 10, 100)

	var messages []db.Message
	result := db.DB.
		Where("sent = ? AND id > ?", true, lastID).
		Order("id ASC").
		Limit(limit).
		Find(&messages)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve messages",
		})
	}

	return c.JSON(fiber.Map{
		"last_id": lastID,
		"limit":   limit,
		"data":    messages,
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
