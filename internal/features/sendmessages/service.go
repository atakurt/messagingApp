package sendmessages

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	httpClient "github.com/atakurt/messagingApp/internal/infrastructure/http"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	redisClient "github.com/atakurt/messagingApp/internal/infrastructure/redis"
	"github.com/atakurt/messagingApp/internal/infrastructure/repository"
	"go.uber.org/zap"
)

type MessageServiceInterface interface {
	ProcessUnsentMessages(ctx context.Context)
}

type MessageService struct {
	repository  *repository.MessageRepository
	httpClient  httpClient.Client
	redisClient redisClient.Client
}

func NewService(repository *repository.MessageRepository, httpClient httpClient.Client, redisClient redisClient.Client) *MessageService {
	return &MessageService{
		repository:  repository,
		httpClient:  httpClient,
		redisClient: redisClient,
	}
}

func (s *MessageService) ProcessUnsentMessages(ctx context.Context) {
	tx, err := s.beginTransaction()
	if err != nil {
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	messages, err := s.fetchUnsentMessages()
	if err != nil {
		return
	}

	logger.Log.Info("Found unsent messages", zap.Int("count", len(messages)))
	if len(messages) == 0 {
		return
	}

	// Process messages concurrently
	processedCount := s.processMessagesConcurrently(ctx, tx, messages)

	logger.Log.Info("Processed messages", zap.Int("count", processedCount))

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Failed to commit transaction", zap.Error(err))
	}
}

func (s *MessageService) beginTransaction() (*db.Transaction, error) {
	tx := db.DB.Begin()
	if tx.Error != nil {
		logger.Log.Error("Failed to begin transaction", zap.Error(tx.Error))
		return nil, tx.Error
	}
	return tx, nil
}

func (s *MessageService) fetchUnsentMessages() ([]db.Message, error) {
	messages, err := s.repository.GetUnsentMessages(config.Cfg.Scheduler.BatchSize)
	if err != nil {
		logger.Log.Error("Failed to select unsent messages with locking", zap.Error(err))
		return nil, err
	}
	return messages, nil
}

// processMessagesConcurrently processes messages in parallel using goroutines
func (s *MessageService) processMessagesConcurrently(ctx context.Context, tx *db.Transaction, messages []db.Message) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	processedMessages := 0

	// Configure concurrency limits
	maxConcurrent := config.Cfg.Scheduler.MaxConcurrent
	semaphore := make(chan struct{}, maxConcurrent)

	for _, msg := range messages {
		wg.Add(1)
		semaphore <- struct{}{}

		msgCopy := msg

		go func(msg db.Message) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			// Process the message and track if successful
			if s.processMessage(ctx, tx, &msg, &mu) {
				mu.Lock()
				processedMessages++
				mu.Unlock()
			}
		}(msgCopy)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	return processedMessages
}

func (s *MessageService) processMessage(ctx context.Context, tx *db.Transaction, msg *db.Message, mu *sync.Mutex) bool {
	redisKey := "message:" + strconv.Itoa(int(msg.ID))

	if !s.canProcessMessage(ctx, redisKey, msg.ID) {
		return false
	}

	now := time.Now()
	mu.Lock()
	err := s.repository.MarkMessageInProcess(msg, now)
	mu.Unlock()
	if err != nil {
		logger.Log.Error("Failed to mark message in process", zap.Uint("messageID", msg.ID), zap.Error(err))
		return false
	}

	hookResp, err := s.sendMessageToWebhook(tx, msg, mu)
	if err != nil {
		return false
	}

	return s.finalizeMessageProcessing(ctx, msg, hookResp, now, redisKey, mu)
}

func (s *MessageService) canProcessMessage(ctx context.Context, redisKey string, messageID uint) bool {
	// Try to acquire a lock for this message
	lockAcquired, err := s.redisClient.SetNX(ctx, redisKey+":lock", time.Now().String(), time.Minute)
	if err != nil {
		logger.Log.Warn("Failed to acquire lock in Redis", zap.String("key", redisKey), zap.Error(err))
		return false
	}

	if !lockAcquired {
		logger.Log.Warn("Message being processed by another instance, skipping", zap.Uint("messageID", messageID))
		return false
	}

	// Check if message was already processed
	exists, err := s.redisClient.Exists(ctx, redisKey)
	if err != nil {
		logger.Log.Warn("Failed to check Redis", zap.String("key", redisKey), zap.Error(err))
	}
	if exists {
		logger.Log.Warn("Message already processed, skipping", zap.Uint("messageID", messageID))
		return false
	}

	return true
}

func (s *MessageService) sendMessageToWebhook(tx *db.Transaction, msg *db.Message, mu *sync.Mutex) (*HookResponse, error) {
	payload := WebhookPayload{Message: msg.Content, To: msg.PhoneNumber}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		logger.Log.Error("Failed to encode payload to JSON", zap.Error(err))
		mu.Lock()
		s.repository.UpdateMessageAsError(tx, msg, err.Error())
		mu.Unlock()
		return nil, err
	}

	resp, err := s.httpClient.Post(config.Cfg.WebhookUrl, "application/json", buf)
	if err != nil {
		logger.Log.Error("Failed to send message", zap.Error(err))
		mu.Lock()
		// todo retry caching with exponencial backoff
		s.repository.InsertRetry(tx, *msg, err.Error())
		mu.Unlock()
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read webhook response", zap.Error(err))
		mu.Lock()
		s.repository.InsertRetry(tx, *msg, err.Error())
		mu.Unlock()
		return nil, err
	}

	var hookResp HookResponse
	if err := json.Unmarshal(bodyBytes, &hookResp); err != nil {
		logger.Log.Error("Failed to parse webhook response", zap.ByteString("body", bodyBytes), zap.Error(err))
		mu.Lock()
		// todo retry caching with exponencial backoff
		s.repository.InsertRetry(tx, *msg, err.Error())
		mu.Unlock()
		return nil, err
	}

	return &hookResp, nil
}

func (s *MessageService) finalizeMessageProcessing(
	ctx context.Context,
	msg *db.Message,
	hookResp *HookResponse,
	timestamp time.Time,
	redisKey string,
	mu *sync.Mutex,
) bool {
	mu.Lock()
	if err := s.repository.UpdateMessageAsSent(msg, hookResp.MessageID, timestamp); err != nil {
		logger.Log.Error("Failed to update message", zap.Uint("messageID", msg.ID), zap.Error(err))
		mu.Unlock()
		return false
	}
	mu.Unlock()

	// todo retry caching with exponencial backoff
	err := s.redisClient.Set(ctx, redisKey, timestamp.String(), time.Hour)
	if err != nil {
		logger.Log.Warn("Failed to cache message in Redis", zap.String("key", redisKey), zap.Error(err))
	}

	logger.Log.Info("Message sent and cached",
		zap.Uint("messageID", msg.ID),
		zap.String("to", msg.PhoneNumber),
		zap.String("messageId", hookResp.MessageID))

	return true
}
