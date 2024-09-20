package logger

import (
	"strings"

	"go.uber.org/zap"
)

func Init(environment string) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	if strings.EqualFold(environment, "development") {
		logger, err = zap.NewDevelopment()

		if err != nil {
			return err
		}
	}

	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	return nil
}
