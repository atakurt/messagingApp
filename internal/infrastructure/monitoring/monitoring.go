package monitoring

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/gofiber/fiber/v2"
)

type MonitoringInterface interface {
	Readiness(c *fiber.Ctx) error
	Liveness(c *fiber.Ctx) error
}

type MonitoringService struct {
	db    db.DBInterface
	redis redisClient.Client
}

func NewMonitoringService(dbInstance db.DBInterface, redis redisClient.Client) *MonitoringService {
	return &MonitoringService{db: dbInstance, redis: redis}
}

func (s *MonitoringService) Readiness(c *fiber.Ctx) error {
	sqlDB, err := s.db.GetSQLDB()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString("DB connection error")
	}
	if err := sqlDB.Ping(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString("DB ping failed")
	}

	// Redis check
	status := s.redis.Ping(c.Context())
	if status.Err() != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString("Redis ping failed")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (s *MonitoringService) Liveness(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
