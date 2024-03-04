package internal

import "go.uber.org/zap"

var Logger zap.SugaredLogger

func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	Logger = *logger.Sugar()
}
