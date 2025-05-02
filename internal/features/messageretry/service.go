package messageretry

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/atakurt/messagingApp/internal/features/sendmessages"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/db"
	httpClient "github.com/atakurt/messagingApp/internal/infrastructure/http"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/atakurt/messagingApp/internal/infrastructure/repository"
	"github.com/cenkalti/backoff/v5"
	"go.uber.org/zap"
)

type MessageRetryServiceInterface interface {
	ProcessMessageRetries(ctx context.Context)
}

type MessageRetryService struct {
	repository *repository.MessageRepository
	httpClient httpClient.Client
}

func NewService(repository *repository.MessageRepository, httpClient httpClient.Client) *MessageRetryService {
	return &MessageRetryService{
		repository: repository,
		httpClient: httpClient,
	}
}

func (s *MessageRetryService) ProcessMessageRetries(ctx context.Context) {
	tx, err := s.beginTransaction()
	if err != nil {
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	retries, err := s.fetchPendingRetries()
	if err != nil {
		return
	}

	logger.Log.Info("Found message retries", zap.Int("count", len(retries)))
	if len(retries) == 0 {
		return
	}

	// Process retries concurrently
	processedCount := s.processRetriesConcurrently(ctx, tx, retries)

	logger.Log.Info("Processed retries", zap.Int("count", processedCount))

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Failed to commit transaction", zap.Error(err))
	}
}

func (s *MessageRetryService) beginTransaction() (*db.Transaction, error) {
	tx := db.DB.Begin()
	if tx.Error != nil {
		logger.Log.Error("Failed to begin transaction", zap.Error(tx.Error))
		return nil, tx.Error
	}
	return tx, nil
}

func (s *MessageRetryService) fetchPendingRetries() ([]db.MessageRetry, error) {
	retries, err := s.repository.GetMessageRetries(config.Cfg.Scheduler.BatchSize)
	if err != nil {
		logger.Log.Error("Failed to select message retries with locking", zap.Error(err))
		return nil, err
	}
	return retries, nil
}

// processRetriesConcurrently processes retries in parallel using goroutines
func (s *MessageRetryService) processRetriesConcurrently(ctx context.Context, tx *db.Transaction, retries []db.MessageRetry) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	processedRetries := 0

	// Configure concurrency limits
	maxConcurrent := config.Cfg.Scheduler.MaxRetryConcurrent
	semaphore := make(chan struct{}, maxConcurrent)

	for _, retry := range retries {
		wg.Add(1)
		semaphore <- struct{}{}

		retryCopy := retry

		go func(retry db.MessageRetry) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			// Process the retry and track if successful
			if s.processRetry(ctx, tx, &retry, &mu) {
				mu.Lock()
				processedRetries++
				mu.Unlock()
			}
		}(retryCopy)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	return processedRetries
}

func (s *MessageRetryService) processRetry(ctx context.Context, tx *db.Transaction, retry *db.MessageRetry, mu *sync.Mutex) bool {
	// Increment retry count
	newRetryCount := retry.RetryCount + 1
	if newRetryCount > 5 {
		msg := db.Message{
			ID:          retry.OriginalMessageID,
			PhoneNumber: retry.PhoneNumber,
			Content:     retry.Content,
		}

		mu.Lock()
		err := s.repository.MoveToDeadLetter(tx, msg, retry.LastError)
		mu.Unlock()

		if err != nil {
			logger.Log.Error("Failed to move message to dead letter queue",
				zap.Uint("retryID", retry.ID),
				zap.Uint("originalMessageID", retry.OriginalMessageID),
				zap.Error(err))
			return false
		}

		logger.Log.Info("Message moved to dead letter queue after max retries",
			zap.Uint("originalMessageID", retry.OriginalMessageID),
			zap.Int("retryCount", newRetryCount))
		return true
	}

	// Create an exponential backoff configuration
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 1 * time.Second
	expBackoff.MaxInterval = 5 * time.Second
	expBackoff.Multiplier = 2.0
	expBackoff.RandomizationFactor = 0.2 // jitter

	// Calculate the backoff duration based on the current retry count
	var backoffDuration time.Duration
	for i := 0; i < retry.RetryCount; i++ {
		backoffDuration = expBackoff.NextBackOff()
	}

	logger.Log.Info("Applying exponential backoff before retry",
		zap.Uint("retryID", retry.ID),
		zap.Int("retryCount", retry.RetryCount),
		zap.Duration("backoffDuration", backoffDuration))

	select {
	case <-ctx.Done():
		return false
	case <-time.After(backoffDuration):
		// Continue with retry
	}

	hookResp, err := s.sendMessageToWebhook(retry)
	if err != nil {
		mu.Lock()
		updateErr := s.repository.UpdateRetryCount(tx, retry.ID, newRetryCount, err.Error())
		mu.Unlock()

		if updateErr != nil {
			logger.Log.Error("Failed to update retry count",
				zap.Uint("retryID", retry.ID),
				zap.Error(updateErr))
		}

		logger.Log.Warn("Retry attempt failed",
			zap.Uint("retryID", retry.ID),
			zap.Int("retryCount", newRetryCount),
			zap.Error(err))
		return false
	}

	// Message sent successfully, update the original message
	msg := &db.Message{ID: retry.OriginalMessageID}
	now := time.Now()

	mu.Lock()
	err = s.repository.UpdateMessageAsSent(msg, hookResp.MessageID, now)
	mu.Unlock()

	if err != nil {
		logger.Log.Error("Failed to update original message after successful retry",
			zap.Uint("messageID", retry.OriginalMessageID),
			zap.Error(err))
		return false
	}

	logger.Log.Info("Message retry successful",
		zap.Uint("originalMessageID", retry.OriginalMessageID),
		zap.Int("retryCount", newRetryCount),
		zap.String("messageId", hookResp.MessageID))

	return true
}

func (s *MessageRetryService) sendMessageToWebhook(retry *db.MessageRetry) (*sendmessages.HookResponse, error) {
	payload := sendmessages.WebhookPayload{Message: retry.Content, To: retry.PhoneNumber}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Post(config.Cfg.WebhookUrl, "application/json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var hookResp sendmessages.HookResponse
	if err := json.Unmarshal(bodyBytes, &hookResp); err != nil {
		return nil, err
	}

	return &hookResp, nil
}
