package logrus

import (
	"testing"
	"time"
)

func TestRotatingLogger(t *testing.T) {
	ConfigRotatingFileLogger(&Config{
		logPath:      "/Users/wayne/Work/CENTNET/Code/Projects/asr-service/common/log/logrus/",
		logFile:      "applog",
		level:        DebugLevel,
		maxAge:       time.Hour * 24 * 7,
		rotationTime: time.Hour * 24,
	})

	Logger.Debug("abcdefghijklmn")
}
