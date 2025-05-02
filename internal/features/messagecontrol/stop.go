package messagecontrol

import (
	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler"
	"github.com/gofiber/fiber/v2"
)

// StopHandler godoc
// @Summary Stop automatic message sending
// @Tags Scheduler
// @Success 200 {string} string "Scheduler stopped"
// @Router /stop [post]
func StopHandler(ctx *fiber.Ctx, redisClient redisClient.Client) error {
	err := scheduler.PublishCommand(ctx.Context(), redisClient, "stop")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to publish stop command",
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "Stop command sent to all scheduler instances",
	})
}
