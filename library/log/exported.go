package log

func Debug(args ...interface{}) {
	_logger.Debug(args...)
}

func Info(args ...interface{}) {
	_logger.Info(args...)
}

func Warn(args ...interface{}) {
	_logger.Warn(args...)
}

func Error(args ...interface{}) {
	_logger.Error(args...)
}

//func DPanic(args ...interface{}) {
//	_logger.DPanic(args...)
//}

func Panic(args ...interface{}) {
	_logger.Panic(args...)
}

func Fatal(args ...interface{}) {
	_logger.Fatal(args...)
}

func Debugf(template string, args ...interface{}) {
	_logger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	_logger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	_logger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	_logger.Errorf(template, args...)
}

//func DPanicf(template string, args ...interface{}) {
//	_logger.DPanicf(template, args...)
//}

func Panicf(template string, args ...interface{}) {
	_logger.Panicf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	_logger.Fatalf(template, args...)
}

func Init(c *Config) {
	_logger = NewLogger(c)
}

func DefaultLogger() *Logger {
	return _logger
}
