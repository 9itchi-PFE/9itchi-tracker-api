package utils

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogWrapperObj struct {
	Logger *zap.SugaredLogger
}

var Logger = LogWrapperObj{
	Logger: initLogger(),
}

func initLogger() *zap.SugaredLogger {
	var logger *zap.Logger
	var err error

	if gin.IsDebugging() {
		logger, err = zap.NewDevelopment()
	} else {
		// In production, log Info level and above.
		config := zap.NewProductionConfig()
		config.Level.SetLevel(zap.InfoLevel)
		config.EncoderConfig.TimeKey = "date_time"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// Disable sampling to ensure safe concurrent calls.
		config.Sampling = nil
		logger, err = config.Build()
		// Attach a namespace to production logs.
		logger = logger.With(zap.Namespace("app"))
	}

	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	zap.ReplaceGlobals(logger)
	return logger.Sugar()
}

func (logWrapper LogWrapperObj) Info(message string, fields ...zap.Field) {
	logWrapper.Logger.Desugar().Info(message, fields...)
}

func (logWrapper LogWrapperObj) Warn(message string, fields ...zap.Field) {
	logWrapper.Logger.Desugar().Warn(message, fields...)
}

func (logWrapper LogWrapperObj) Error(message string, fields ...zap.Field) {
	logWrapper.Logger.Desugar().Error(message, fields...)
}

func (logWrapper LogWrapperObj) Fatal(message string, fields ...zap.Field) {
	logWrapper.Logger.Desugar().Fatal(message, fields...)
}
