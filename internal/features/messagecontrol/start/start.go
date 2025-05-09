package start

import (
	"github.com/atakurt/messagingApp/internal/features/messagecontrol"
	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler"
	"github.com/gofiber/fiber/v2"
)

// StartHandler godoc
// @Summary Start automatic message sending
// @Tags Scheduler
// @Success 200 {string} string "Scheduler started"
// @Router /start [post]
func StartHandler(ctx *fiber.Ctx, redisClient redisClient.Client) error {
	err := scheduler.PublishCommand(ctx.Context(), redisClient, "start")
	if err != nil {
		schedulerErr := &messagecontrol.SchedulerError{
			Operation: "start",
			Err:       err,
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": schedulerErr.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "Start command sent to all scheduler instances",
	})
}
