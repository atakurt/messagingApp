package main

import (
	"fmt"
	"github.com/atakurt/messagingApp/internal/features/messagecontrol/commandlistener"
	"github.com/atakurt/messagingApp/internal/features/messagecontrol/list_sent"
	"github.com/atakurt/messagingApp/internal/features/messagecontrol/start"
	"github.com/atakurt/messagingApp/internal/features/messagecontrol/stop"
	"github.com/atakurt/messagingApp/internal/infrastructure/scheduler/retry"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/atakurt/messagingApp/docs"
	"github.com/atakurt/messagingApp/internal/features/messageretry"
	"github.com/atakurt/messagingApp/internal/features/sendmessages"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	httpClient "github.com/atakurt/messagingApp/internal/infrastructure/http"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/infrastructure/monitoring"
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
// @BasePath /
// @contact.name  Dev Team
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger.Init()
	defer logger.Log.Sync()

	config.Init()
	db.Init()

	listenShutdownSignal(cancel)

	gormDB := db.DB.GetDB()

	// Create the message service
	messageRepository := repository.NewMessageRepository(gormDB)
	client := httpClient.NewHttpClient()

	redisClient := redis.NewClient(ctx, goRedis.NewClient(&goRedis.Options{
		Addr: config.Cfg.Redis.Addr,
	}))

	messageService := sendmessages.NewService(messageRepository, client, redisClient)
	messageRetryService := messageretry.NewService(messageRepository, client)

	mainScheduler := scheduler.NewScheduler(messageService, redisClient)
	retryScheduler := retry.NewRetryScheduler(messageRetryService, redisClient, config.Cfg)
	commandListenr := commandlistener.NewCommandListener(redisClient, mainScheduler)

	go commandListenr.Listen(ctx)

	// Start the schedulers
	mainScheduler.Start(ctx)
	retryScheduler.Start(ctx)

	app := fiber.New()

	setupRoutes(app, redisClient, messageRepository)

	listen(app)

	<-ctx.Done()

	shutdown(app, redisClient, mainScheduler, retryScheduler)
}

func listenShutdownSignal(cancel context.CancelFunc) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Log.Info("Shutdown signal received")
		cancel()
	}()
}

func listen(app *fiber.App) {
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", config.Cfg.Server.Port)); err != nil {
			logger.Log.Error("Fiber Listen error", zap.Error(err))
		}
	}()
}

func setupRoutes(app *fiber.App, redisClient *redis.RedisClient, messageRepository *repository.MessageRepository) {
	app.Post("/start", func(ctx *fiber.Ctx) error {
		return start.StartHandler(ctx, redisClient)
	})
	app.Post("/stop", func(ctx *fiber.Ctx) error {
		return stop.StopHandler(ctx, redisClient)
	})

	messagecontrolService := list_sent.NewService(messageRepository)
	app.Get("/sent-messages", func(ctx *fiber.Ctx) error {
		return messagecontrolService.ListSentMessages(ctx)
	})

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", fiber.StatusFound)
	})

	// monitoring
	monitoringService := monitoring.NewMonitoringService(db.DB, redisClient)
	app.Get("/ready", func(c *fiber.Ctx) error {
		return monitoringService.Readiness(c)
	})

	app.Get("/live", func(c *fiber.Ctx) error {
		return monitoringService.Liveness(c)
	})
}

func shutdown(app *fiber.App, redisClient *redis.RedisClient, scheduler *scheduler.Scheduler, retryScheduler *retry.RetryScheduler) {
	logger.Log.Info("Shutting down Fiber app")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	shutdownErr := make(chan error, 1)

	go func() {
		scheduler.Stop(shutdownCtx)
		retryScheduler.Stop(shutdownCtx)

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			shutdownErr <- fmt.Errorf("fiber shutdown error: %w", err)
			return
		}

		if err := redisClient.Close(shutdownCtx); err != nil {
			shutdownErr <- fmt.Errorf("redis close error: %w", err)
			return
		}

		sqlDB, _ := db.DB.GetSQLDB()
		if err := sqlDB.Close(); err != nil {
			shutdownErr <- fmt.Errorf("postgres close error: %w", err)
			return
		}

		shutdownErr <- nil
	}()

	select {
	case err := <-shutdownErr:
		if err != nil {
			logger.Log.Error("Shutdown error", zap.Error(err))
		} else {
			logger.Log.Info("Shutdown complete")
		}
	case <-shutdownCtx.Done():
		logger.Log.Error("Shutdown timed out")
	}
}
