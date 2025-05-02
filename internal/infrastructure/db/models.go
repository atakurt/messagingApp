package db

import "time"

type MessageStatus string

const (
	StatusPending    MessageStatus = "pending"
	StatusProcessing MessageStatus = "processing"
	StatusDone       MessageStatus = "done"
	StatusError      MessageStatus = "error"
)

type Message struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	PhoneNumber string        `json:"phone_number"`
	Content     string        `json:"content"`
	Status      MessageStatus `gorm:"default:pending" json:"status"`
	MessageID   string        `json:"message_id,omitempty"`
	LastError   string        `json:"last_error,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	ProcessedAt time.Time     `json:"processed_at,omitempty"`
	SentAt      time.Time     `json:"sent_at,omitempty"`
}

type MessageRetry struct {
	ID                uint   `gorm:"primaryKey"`
	OriginalMessageID uint   `gorm:"not null;index"`
	PhoneNumber       string `gorm:"not null"`
	Content           string `gorm:"not null"`
	RetryCount        int    `gorm:"not null;default:0"`
	LastError         string
	CreatedAt         time.Time
}

type MessageDeadLetter struct {
	ID                uint   `gorm:"primaryKey"`
	OriginalMessageID uint   `gorm:"not null;index"`
	PhoneNumber       string `gorm:"not null"`
	Content           string `gorm:"not null"`
	LastError         string
	FailedAt          time.Time `gorm:"autoCreateTime"`
}
