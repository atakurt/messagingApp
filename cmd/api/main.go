package main

import (
	_ "github.com/atakurt/messagingApp/docs"
	"github.com/atakurt/messagingApp/internal/features/messagecontrol"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Message Sender API
// @version 1.0
// @description Auto message scheduler
// @host localhost:8080
// @BasePath /
// @contact.name  Dev Team
func main() {
	logger.Init()
	defer logger.Log.Sync()

	config.Init()

	db.Init()
	redis.Init()
	scheduler.Start()

	app := fiber.New()

	app.Post("/start", messagecontrol.StartHandler)
	app.Post("/stop", messagecontrol.StopHandler)
	app.Get("/sent-messages", messagecontrol.ListSentMessages)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", fiber.StatusFound)
	})

	app.Listen(":8080")
}
