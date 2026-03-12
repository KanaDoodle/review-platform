package logger

import "go.uber.org/zap"

func New() *zap.Logger {
	log, _ := zap.NewProduction()
	return log
}

func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Err(err error) zap.Field {
	return zap.Error(err)
}