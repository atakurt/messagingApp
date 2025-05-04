package config

import (
	"os"
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Port int
	}
	Scheduler struct {
		Enabled            bool
		Interval           time.Duration
		BatchSize          int
		MaxConcurrent      int
		MaxRetryConcurrent int
	}
	Database struct {
		DSN string
	}
	Redis struct {
		Addr string
	}

	Http struct {
		Timeout               time.Duration
		MaxIdleConns          int
		IdleConnTimeout       time.Duration
		TlsHandshakeTimeout   time.Duration
		ExpectContinueTimeout time.Duration
	}

	WebhookUrl string
}

var Cfg Config

func Init() {
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs"
	}
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("APP")
	viper.SetDefault("scheduler.enabled", true)
	viper.SetDefault("scheduler.maxConcurrent", 1)
	viper.SetDefault("scheduler.maxRetryConcurrent", 1)
	viper.AutomaticEnv()

	viper.BindEnv("DATABASE_DSN")
	viper.BindEnv("REDIS_ADDR")
	viper.BindEnv("WEBHOOK_URL")
	viper.BindEnv("SCHEDULER_INTERVAL")
	viper.BindEnv("SCHEDULER_BATCHSIZE")

	if err := viper.ReadInConfig(); err != nil {
		logger.Log.Fatal("Error reading config", zap.Error(err))
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		logger.Log.Fatal("Unable to decode config into struct", zap.Error(err))
	}

	if envDSN := viper.GetString("DATABASE_DSN"); envDSN != "" {
		Cfg.Database.DSN = envDSN
		logger.Log.Info("database.dsn overridden by env", zap.String("dsn", envDSN))
	}
	if envRedis := viper.GetString("REDIS_ADDR"); envRedis != "" {
		Cfg.Redis.Addr = envRedis
		logger.Log.Info("redis.addr overridden by env", zap.String("redis.addr", envRedis))
	}

	if webhookUrl := viper.GetString("WEBHOOK_URL"); webhookUrl != "" {
		Cfg.WebhookUrl = webhookUrl
		logger.Log.Info("webhookUrl overridden by env", zap.String("webhookUrl", webhookUrl))
	}

	if interval := viper.GetDuration("SCHEDULER_INTERVAL"); interval != 0 {
		Cfg.Scheduler.Interval = interval
		logger.Log.Info("scheduler.interval overridden", zap.Duration("interval", interval))
	}

	if batchSize := viper.GetInt("SCHEDULER_BATCHSIZE"); batchSize != 0 {
		Cfg.Scheduler.BatchSize = batchSize
		logger.Log.Info("scheduler.batchsize overridden", zap.Int("batchsize", batchSize))
	}

	if maxConcurrent := viper.GetInt("SCHEDULER_MAX_CONCURRENT"); maxConcurrent != 0 {
		Cfg.Scheduler.MaxConcurrent = maxConcurrent
		logger.Log.Info("scheduler.maxConcurrent overridden", zap.Int("maxConcurrent", maxConcurrent))
	}

	logger.Log.Info("scheduler.enabled", zap.Bool("enabled", Cfg.Scheduler.Enabled))

	logger.Log.Info("Loaded config file", zap.String("file", viper.ConfigFileUsed()))
	logger.Log.Info("Loaded config value", zap.Any("cfg", Cfg))

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Log.Info("Config file changed", zap.String("file", e.Name))
		viper.Unmarshal(&Cfg)
	})
}
