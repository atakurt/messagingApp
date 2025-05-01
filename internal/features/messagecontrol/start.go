package messagecontrol

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler"
	"github.com/gofiber/fiber/v2"
)

// StartHandler godoc
// @Summary Start automatic message sending
// @Tags Scheduler
// @Success 200 {string} string "Scheduler started"
// @Router /start [post]
func StartHandler(c *fiber.Ctx) error {
	scheduler.Start()
	return c.SendString("Scheduler started")
}
