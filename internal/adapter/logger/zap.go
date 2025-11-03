package logger

import (
	"fmt"
	"restaurant/internal/adapter/config"

	"go.uber.org/zap"
)

// setZapLogger sets global zap global logger.
func setZapLogger(appConfig *config.AppConfig) error {
	switch appConfig.Environment {
	case config.Production:
		loggerConfig := zap.NewProductionConfig()
		loggerConfig.DisableStacktrace = true

		logger, err := loggerConfig.Build()
		if err != nil {
			return err
		}
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
