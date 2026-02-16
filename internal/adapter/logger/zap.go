package logger

import (
	"fmt"
	"restaurant/internal/adapter/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SetZapLogger sets global zap logger.
func SetZapLogger(appConfig *config.AppConfig) error {
	switch appConfig.Environment {
	case config.Production:
		lumberLogger := &lumberjack.Logger{
			Filename:   "./logs",
			MaxSize:    1,
			MaxBackups: 3,
			MaxAge:     2,
			Compress:   true,
		}

		loggerConfig := zap.NewProductionConfig()
		loggerConfig.DisableStacktrace = true
		loggerConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)

		writeSyncer := zapcore.AddSync(lumberLogger)
		encoder := zapcore.NewJSONEncoder(loggerConfig.EncoderConfig)

		core := zapcore.NewCore(
			encoder,
			writeSyncer,
			loggerConfig.Level,
		)

		zap.ReplaceGlobals(zap.New(core))

	case config.Development:
		loggerConfig := zap.NewDevelopmentConfig()
		loggerConfig.DisableStacktrace = true

		logger, err := loggerConfig.Build()
		if err != nil {
			return err
		}
		zap.ReplaceGlobals(logger)

	default:
		return fmt.Errorf("unknown environment: %s", appConfig.Environment)
	}

	return nil
}
