package sendmessages

import (
	"bytes"
	"encoding/json"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"gorm.io/gorm/clause"

	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type WebhookPayload struct {
	Message string `json:"message"`
	To      string `json:"to"`
}

func SendUnsentMessages() {
	tx := db.DB.Begin()
	if tx.Error != nil {
		logger.Log.Error("failed to begin transaction", zap.Error(tx.Error))
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var messages []db.Message
	err := tx.Clauses(
		clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"},
	).Limit(config.Cfg.Scheduler.BatchSize).
		Where("sent = false").
		Find(&messages).Error

	if err != nil {
		logger.Log.Error("Failed to select unsent messages with locking", zap.Error(err))
		return
	}

	logger.Log.Info("Found unsent messages", zap.Int("count", len(messages)))

	for _, msg := range messages {
		redisKey := "message:" + strconv.Itoa(int(msg.ID))

		exists, err := redisClient.Rdb.Exists(redisClient.Ctx, redisKey).Result()
		if err != nil {
			logger.Log.Warn("Failed to check Redis", zap.String("key", redisKey), zap.Error(err))
		}
		if exists == 1 {
			logger.Log.Warn("Message already processed, skipping", zap.Uint("messageID", msg.ID))
			continue
		}

		payload := WebhookPayload{
			Message: msg.Content,
			To:      msg.PhoneNumber,
		}

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			logger.Log.Error("Failed to encode payload to JSON", zap.Error(err))
			return
		}

		resp, err := http.Post(config.Cfg.WebhookUrl, "application/json", buf)
		if err != nil {
			logger.Log.Error("Failed to send message", zap.Error(err))
			continue
		}
		resp.Body.Close()

		now := time.Now()
		if err := tx.Model(&msg).Updates(db.Message{
			Sent:   true,
			SentAt: &now,
		}).Error; err != nil {
			logger.Log.Error("Failed to update message after sending", zap.Uint("messageID", msg.ID), zap.Error(err))
			return
		}

		err = redisClient.Rdb.Set(redisClient.Ctx, redisKey, now.String(), 1*time.Hour).Err()
		if err != nil {
			logger.Log.Warn("Failed to cache message in Redis", zap.String("key", redisKey), zap.Error(err))
		}

		logger.Log.Info("Message sent and cached", zap.Uint("messageID", msg.ID), zap.String("to", msg.PhoneNumber))
	}

	// Commit only after all messages processed
	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Failed to commit transaction", zap.Error(err))
	}
}
