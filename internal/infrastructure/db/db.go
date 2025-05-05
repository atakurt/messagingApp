package db

import (
	"database/sql"
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

//go:generate mockgen -destination=../../mocks/mock_db_client.go -package=mocks github.com/atakurt/messagingApp/internal/infrastructure/db DBInterface
type DBInterface interface {
	GetSQLDB() (*sql.DB, error)
	GetDB() *gorm.DB
	Begin() *gorm.DB
}

type GormDB struct {
	*gorm.DB
}

func (g *GormDB) GetSQLDB() (*sql.DB, error) {
	return g.DB.DB()
}

func (g *GormDB) GetDB() *gorm.DB {
	return g.DB
}

func (g *GormDB) Begin() *gorm.DB {
	return g.DB.Begin()
}

var DB DBInterface

type Transaction = gorm.DB

func Init() {
	var err error
	zl := zapgorm2.New(logger.Log)
	zl.SlowThreshold = 100 * time.Millisecond
	zl.LogLevel = gormLogger.Warn
	zl.IgnoreRecordNotFoundError = true

	gormDB, err := gorm.Open(postgres.Open(config.Cfg.Database.DSN), &gorm.Config{
		PrepareStmt:            false,
		SkipDefaultTransaction: true,
		Logger:                 zl,
	})

	if err != nil {
		logger.Log.Fatal("Failed to connect to DB", zap.Error(err))
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Log.Fatal("Failed to get DB from GORM", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(2 * time.Minute)

	// Set the global DB instance
	DB = &GormDB{DB: gormDB}
}
