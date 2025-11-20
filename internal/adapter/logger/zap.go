package logger

import (
	"fmt"
	"restaurant/internal/adapter/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// setZapLogger sets global zap global logger.
func setZapLogger(appConfig *config.AppConfig) error {
	switch appConfig.Environment {
	case config.Production:
		lumberLogger := &lumberjack.Logger{
			Filename:   "./logs",
			MaxSize:    30,
			MaxBackups: 3,
			MaxAge:     2,
			Compress:   true,
		}

		loggerConfig := zap.NewProductionConfig()
		loggerConfig.DisableStacktrace = true

		writeSyncer := zapcore.AddSync(lumberLogger)
		encoder := zapcore.NewJSONEncoder(loggerConfig.EncoderConfig)

		core := zapcore.NewCore(
			encoder,
			writeSyncer,
			loggerConfig.Level,
		)

		logger := zap.New(core, zap.AddCaller())
		zap.ReplaceGlobals(logger)

	case config.Development:
		loggerConfig := zap.NewDevelopmentConfig()
		loggerConfig.DisableStacktrace = true

		logger, err := loggerConfig.Build()
		if err != nil {
			return err
		}
		zap.ReplaceGlobals(logger)

	default:
		return fmt.Errorf("unknown environment %s", appConfig.Environment)
	}

	return nil
}
