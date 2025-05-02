package main

import (
	_ "github.com/atakurt/messagingApp/docs"
	"github.com/atakurt/messagingApp/internal/features/messagecontrol"
	"github.com/atakurt/messagingApp/internal/features/messageretry"
	"github.com/atakurt/messagingApp/internal/features/sendmessages"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	httpClient "github.com/atakurt/messagingApp/internal/infrastructure/http"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/atakurt/messagingApp/internal/infrastructure/repository"
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"golang.org/x/net/context"
)

// @title Message Sender API
// @version 1.0
// @description Auto message scheduler
// @host localhost:8080
// @BasePath /
// @contact.name  Dev Team
func main() {
	defaultCtx := context.Background()
	logger.Init()
	defer logger.Log.Sync()

	config.Init()
	db.Init()

	// Create the message service
	messageRepository := repository.NewMessageRepository(db.DB)
	client := httpClient.NewHttpClient()

	redisClient := redis.NewClient(defaultCtx, goRedis.NewClient(&goRedis.Options{
		Addr: config.Cfg.Redis.Addr,
	}))

	messageService := sendmessages.NewService(messageRepository, client, redisClient)
	messageRetryService := messageretry.NewService(messageRepository, client)

	mainScheduler := scheduler.NewScheduler(messageService, redisClient)
	retryScheduler := scheduler.NewRetryScheduler(messageRetryService, redisClient, config.Cfg)

	// Start the schedulers
	mainScheduler.Start(defaultCtx)
	retryScheduler.Start(defaultCtx)

	app := fiber.New()

	app.Post("/start", func(ctx *fiber.Ctx) error {
		return messagecontrol.StartHandler(ctx, redisClient)
	})
	app.Post("/stop", func(ctx *fiber.Ctx) error {
		return messagecontrol.StopHandler(ctx, redisClient)
	})

	messagecontrolService := messagecontrol.NewService(messageRepository)
	app.Get("/sent-messages", func(ctx *fiber.Ctx) error {
		return messagecontrolService.ListSentMessages(ctx)
	})

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", fiber.StatusFound)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Listen(":8080")
}
