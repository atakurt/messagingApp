package db

import (
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(postgres.Open(config.Cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to DB", zap.Error(err))
	}

	//if err := DB.AutoMigrate(&Message{}); err != nil {
	//	logger.Log.Fatal("Failed to auto-migrate DB", zap.Error(err))
	//}
}

type Message struct {
	ID          uint       `gorm:"primaryKey"`
	PhoneNumber string     `gorm:"not null;size:20"`
	Content     string     `gorm:"not null;size:160"`
	Sent        bool       `gorm:"not null;default:false;index"`
	SentAt      *time.Time `gorm:"column:sent_at"`
	MessageID   string     `gorm:"column:message_id;size:255"`
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
}
