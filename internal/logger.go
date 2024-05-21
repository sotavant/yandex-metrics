package internal

import "go.uber.org/zap"

var Logger zap.SugaredLogger

func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer func(logger *zap.Logger) {
		err = logger.Sync()
	}(logger)

	if err != nil {
		panic(err)
	}
	Logger = *logger.Sugar()
}
