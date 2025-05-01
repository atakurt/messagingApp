package messagecontrol

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	"github.com/gofiber/fiber/v2"
)

// ListSentMessages godoc
// @Summary      List sent messages
// @Description  Retrieves all messages that have been marked as sent
// @Tags         Messages
// @Produce      json
// @Success      200  {array}  db.Message
// @Failure      500  {object}  map[string]string
// @Router       /sent-messages [get]
func ListSentMessages(c *fiber.Ctx) error {
	var messages []db.Message
	result := db.DB.Where("sent = ?", true).Find(&messages)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve messages",
		})
	}

	return c.JSON(messages)
}
