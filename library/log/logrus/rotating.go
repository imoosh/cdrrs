package logrus

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	PanicLevel = logrus.PanicLevel
	FatalLevel = logrus.FatalLevel
	ErrorLevel = logrus.ErrorLevel
	WarnLevel  = logrus.WarnLevel
	InfoLevel  = logrus.InfoLevel
	DebugLevel = logrus.DebugLevel
	TraceLevel = logrus.TraceLevel
)

type Config struct {
	logPath      string
	logFile      string
	level        logrus.Level
	maxAge       time.Duration
	rotationTime time.Duration
}

var DefaultConfig = Config{
	logPath:      "./",
	logFile:      "applog",
	level:        DebugLevel,
	maxAge:       time.Hour * 24 * 7,
	rotationTime: time.Hour * 24,
}

var Logger = logrus.New()

//func ConfigRotatingFileLogger(c *Conf) *logrus.Logger {
func ConfigRotatingFileLogger(c *Config) {
	if c != nil {
		c = &DefaultConfig
	}

	var fileName string
	if strings.HasSuffix(c.logPath, "/") {
		fileName = path.Join(c.logPath, c.logFile)
	} else {
		fileName = path.Join(c.logPath, "/", c.logFile)
	}

	// 创建logger
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(c.level)
	Logger.SetFormatter(&logrus.TextFormatter{CallerPrettyfier: callerInfo})
	Logger.SetReportCaller(true)

	// io.Writer
	logWriter, err := rotatelogs.New(
		fileName+".%Y%m%d.log",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(c.maxAge),
		rotatelogs.WithRotationTime(c.rotationTime),
	)
	if err != nil {
		panic(err)
	}

	// 为不通级别的日志设置不通的输出目的
	writerMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	// 创建LfsHook
	lfHook := lfshook.NewHook(writerMap, &logrus.JSONFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		CallerPrettyfier: callerInfo,
	})

	// 注册hook
	Logger.AddHook(lfHook)
}

// callerInfo get short function name & file name
func callerInfo(c *runtime.Frame) (function string, file string) {
	n := strings.LastIndexByte(c.File, '/')
	fileVal := fmt.Sprintf("%s:%d", c.File[n+1:], c.Line)
	n = strings.LastIndexByte(c.Function, '/')
	funcVal := c.Function[n+1:]
	return funcVal, fileVal
}
