package repository

import (
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate mockgen -destination=../../mocks/mock_repository.go -package=mocks github.com/atakurt/messagingApp/internal/infrastructure/repository MessageRepositoryInterface
type MessageRepositoryInterface interface {
	GetUnsentMessages(tx *gorm.DB, limit int) ([]db.Message, error)
	MarkMessageInProcess(tx *gorm.DB, msg *db.Message, processedAt time.Time) error
	UpdateMessageAsError(tx *gorm.DB, msg *db.Message, errMsg string) error
	GetSentMessages(lastID, limit int) ([]db.Message, error)
	UpdateMessageAsSent(tx *gorm.DB, msg *db.Message, messageID string, sentAt time.Time) error
	InsertRetry(tx *gorm.DB, msg db.Message, errMsg string) error
	GetMessageRetries(tx *gorm.DB, limit int) ([]db.MessageRetry, error)
	UpdateRetryCount(tx *gorm.DB, retryID uint, count int, errMsg string) error
	MoveToDeadLetter(tx *gorm.DB, msg db.Message, errMsg string) error
	GetDB() *gorm.DB
}

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) GetUnsentMessages(tx *gorm.DB, limit int) ([]db.Message, error) {
	var messages []db.Message
	err := tx.Clauses(
		clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"},
	).Limit(limit).
		Where("status = ?", db.StatusPending).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) MarkMessageInProcess(tx *gorm.DB, msg *db.Message, processedAt time.Time) error {
	return tx.Model(msg).Updates(map[string]interface{}{
		"Status":      db.StatusProcessing,
		"ProcessedAt": processedAt,
	}).Error
}

func (r *MessageRepository) UpdateMessageAsError(tx *gorm.DB, msg *db.Message, errMsg string) error {
	update := map[string]interface{}{
		"Status":    db.StatusError,
		"LastError": errMsg,
	}
	return tx.Model(msg).Updates(update).Error
}

func (r *MessageRepository) GetSentMessages(lastID, limit int) ([]db.Message, error) {
	var messages []db.Message
	result := r.db.
		Where("status = ? AND id > ?", db.StatusDone, lastID).
		Order("id ASC").
		Limit(limit).
		Find(&messages)

	return messages, result.Error
}

func (r *MessageRepository) UpdateMessageAsSent(tx *gorm.DB, msg *db.Message, messageID string, sentAt time.Time) error {
	update := map[string]interface{}{
		"Status":    db.StatusDone,
		"SentAt":    sentAt,
		"MessageID": messageID,
	}
	return tx.Model(msg).Updates(update).Error
}

func (r *MessageRepository) InsertRetry(tx *gorm.DB, msg db.Message, errMsg string) error {
	retry := db.MessageRetry{
		OriginalMessageID: msg.ID,
		PhoneNumber:       msg.PhoneNumber,
		Content:           msg.Content,
		RetryCount:        1,
		LastError:         errMsg,
		CreatedAt:         time.Now(),
	}
	return tx.Create(&retry).Error
}

func (r *MessageRepository) GetMessageRetries(tx *gorm.DB, limit int) ([]db.MessageRetry, error) {
	var retries []db.MessageRetry
	err := tx.Clauses(
		clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"},
	).Limit(limit).
		Where("retry_count < ?", 5).
		Find(&retries).Error
	return retries, err
}

func (r *MessageRepository) UpdateRetryCount(tx *gorm.DB, retryID uint, count int, errMsg string) error {
	update := map[string]interface{}{
		"RetryCount": count,
		"LastError":  errMsg,
	}
	return tx.Model(&db.MessageRetry{}).Where("id = ?", retryID).Updates(update).Error
}

func (r *MessageRepository) MoveToDeadLetter(tx *gorm.DB, msg db.Message, errMsg string) error {
	deadLetter := db.MessageDeadLetter{
		OriginalMessageID: msg.ID,
		PhoneNumber:       msg.PhoneNumber,
		Content:           msg.Content,
		LastError:         errMsg,
		FailedAt:          time.Now(),
	}
	return tx.Create(&deadLetter).Error
}

func (r *MessageRepository) GetDB() *gorm.DB {
	return r.db
}
