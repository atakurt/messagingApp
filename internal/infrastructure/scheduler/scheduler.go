package scheduler

import (
	"github.com/atakurt/messagingApp/internal/features/sendmessages"
	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"sync"
	"time"
)

var ticker *time.Ticker
var stopChan chan struct{}
var wg sync.WaitGroup
var running bool

func Start() {

	if !config.Cfg.Scheduler.Enabled {
		logger.Log.Warn("Scheduler is disabled by config")
		return
	}
	if running {
		logger.Log.Warn("Scheduler already running")
		return
	}

	if ticker != nil {
		return
	}
	stopChan = make(chan struct{})
	ticker = time.NewTicker(config.Cfg.Scheduler.Interval)
	running = true

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				logger.Log.Info("Scheduler tick - checking for unsent messages")
				sendmessages.SendUnsentMessages()
			case <-stopChan:
				ticker.Stop()
				ticker = nil
				return
			}
		}
	}()
}

func Stop() {
	if !running {
		logger.Log.Warn("Scheduler is not running")
		return
	}

	if stopChan != nil {
		close(stopChan)
		wg.Wait()
	}
}
