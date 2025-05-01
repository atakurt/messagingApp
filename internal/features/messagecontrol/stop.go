package messagecontrol

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler"
	"github.com/gofiber/fiber/v2"
)

// StopHandler godoc
// @Summary Stop automatic message sending
// @Tags Scheduler
// @Success 200 {string} string "Scheduler stopped"
// @Router /stop [post]
func StopHandler(c *fiber.Ctx) error {
	scheduler.Stop()
	return c.SendString("Scheduler stopped")
}
