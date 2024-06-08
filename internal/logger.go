package internal

import (
	"fmt"

	"go.uber.org/zap"
)

var buildDefaultValue = "N/A"

// Logger Глобальная переменная для инициализированного логера
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

func PrintBuildInfo(version, date, commit string) {
	if version == "" {
		version = buildDefaultValue
	}

	if date == "" {
		date = buildDefaultValue
	}

	if commit == "" {
		commit = buildDefaultValue
	}

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", version, date, commit)
}
