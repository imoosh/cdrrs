package log

import (
    "centnet-cdrrs/common/log/_zap"
	"fmt"
	"go.uber.org/zap"
	"time"
)

var _logger *Logger

func NewLogger(c *Config) *Logger {
	logger := &Logger{
		core: _zap.NewZapLogger(_zap.ZapConfig{
			LogPath:    c.LogPath,
			LogFile:    c.LogFile,
			MaxSize:    c.MaxSize,
			MaxAge:     time.Duration(c.MaxAge) * 24 * time.Hour,
			FileLevel:  _zap.GetLoggerLevel(c.FileLevel),
			ConsLevel:  _zap.GetLoggerLevel(c.ConsoleLevel),
			LocalTime:  true,
			Compress:   false,
			CallerSkip: 2,
		}),
	}

	return logger
}

type Config struct {
	LogPath      string `yaml:"logPath"`
	LogFile      string `yaml:"logFile"`
	FileLevel    string `yaml:"fileLevel"`
	ConsoleLevel string `yaml:"consoleLevel"`
	MaxAge       int    `yaml:"maxAge"`
	MaxSize      int    `yaml:"maxSize"`
}

func (c Config) String() string {
	return fmt.Sprintf(`

Logging:
  - LogPath:          %s
  - LogFile:          %s
  - FileLevel:        %s
  - ConsoleLevel:     %s
  - MaxAge:           %d
  - MaxSize:          %d`,
		c.LogPath, c.LogFile, c.FileLevel, c.ConsoleLevel, c.MaxAge, c.MaxSize)
}

type Logger struct {
	core *zap.Logger
}

func (l *Logger) Debug(args ...interface{}) {
	l.core.Sugar().Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.core.Sugar().Info(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.core.Sugar().Warn(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.core.Sugar().Error(args...)
}

//func (l *Logger) DPanic(args ...interface{}) {
//	l.core.Sugar().DPanic(args...)
//}

func (l *Logger) Panic(args ...interface{}) {
	l.core.Sugar().Panic(args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.core.Sugar().Fatal(args...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.core.Sugar().Debugf(template, args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.core.Sugar().Infof(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.core.Sugar().Warnf(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.core.Sugar().Errorf(template, args...)
}

//func (l *Logger) DPanicf(template string, args ...interface{}) {
//	l.core.Sugar().DPanicf(template, args...)
//}

func (l *Logger) Panicf(template string, args ...interface{}) {
	l.core.Sugar().Panicf(template, args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.core.Sugar().Fatalf(template, args...)
}
