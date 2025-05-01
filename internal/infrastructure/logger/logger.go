package logger

import (
	"go.uber.org/zap"
	"os"
)

var Log *zap.Logger

func Init() {
	var err error
	if os.Getenv("ENV") == "production" {
		Log, err = zap.NewProduction(zap.AddCaller())
	} else {
		Log, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
}
