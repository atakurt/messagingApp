package db

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Transaction = gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(postgres.Open(config.Cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to DB", zap.Error(err))
	}
}
