package internal

import "go.uber.org/zap"

func InitLogger() zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	return *logger.Sugar()
}
