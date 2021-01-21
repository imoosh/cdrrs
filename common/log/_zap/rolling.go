package _zap

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"time"
)

type RollingConfig struct {
	FileName  string
	MaxAge    time.Duration
	MaxSize   int
	LocalTime bool
	Compress  bool
}

func RollingWriter(config RollingConfig) io.Writer {
	return &lumberjack.Logger{
		Filename:  config.FileName,
		MaxSize:   config.MaxSize,
		MaxAge:    int(config.MaxAge.Hours() / 24),
		LocalTime: true,
		Compress:  true,
	}
}
