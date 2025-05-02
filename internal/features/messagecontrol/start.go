package messagecontrol

import (
	"fmt"

	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler"
	"github.com/gofiber/fiber/v2"
)

type SchedulerError struct {
	Operation string
	Err       error
}

func (e *SchedulerError) Error() string {
	return fmt.Sprintf("scheduler operation '%s' failed: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying error
func (e *SchedulerError) Unwrap() error {
	return e.Err
}

// StartHandler godoc
// @Summary Start automatic message sending
// @Tags Scheduler
// @Success 200 {string} string "Scheduler started"
// @Router /start [post]
func StartHandler(ctx *fiber.Ctx, redisClient redisClient.Client) error {
	err := scheduler.PublishCommand(ctx.Context(), redisClient, "start")
	if err != nil {
		schedulerErr := &SchedulerError{
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
