package db

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
	"time"
)

var DB *gorm.DB

type Transaction = gorm.DB

func Init() {
	var err error
	zl := zapgorm2.New(logger.Log)
	zl.SlowThreshold = 100 * time.Millisecond
	zl.LogLevel = gormLogger.Warn
	zl.IgnoreRecordNotFoundError = true
	DB, err = gorm.Open(postgres.Open(config.Cfg.Database.DSN), &gorm.Config{
		PrepareStmt:            false,
		SkipDefaultTransaction: true,
		Logger:                 zl,
	})

	if err != nil {
		logger.Log.Fatal("Failed to connect to DB", zap.Error(err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log.Fatal("Failed to get DB from GORM", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(2 * time.Minute)
}
