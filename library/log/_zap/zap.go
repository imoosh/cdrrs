package _zap

import (
    "fmt"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "os"
    "path"
    "time"
)

func comfortableTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
    enc.AppendString(fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d.%06d",
        t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000))
}

func InitZapLog() {
    config := zap.Config{
        Level:             zap.NewAtomicLevelAt(zap.DebugLevel), // 默认为DebugLevel
        Development:       true,                                 // 是否为开发环境，如果是，对DPanicLevel进行堆栈跟踪
        DisableCaller:     false,                                // 是否禁用调用函数的文件名和行号来注释日志。默认进行注释日志
        DisableStacktrace: false,                                // 是否禁用堆栈跟踪捕获。默认对Warn级别以上和生产error级别以上进行堆栈跟踪
        Sampling:          nil,                                  //
        Encoding:          "json",                               // 编码类型，目前两种json和console【按照空格隔开】，常用json
        EncoderConfig: zapcore.EncoderConfig{ // 生成格式配置
            MessageKey:     "msg",
            LevelKey:       "lv",
            TimeKey:        "t",
            NameKey:        "logger",
            CallerKey:      "caller",
            StacktraceKey:  "trace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     comfortableTimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
            EncodeName:     nil,
        },
        OutputPaths:      []string{"/tmp/_zap.log", "stdout"}, // 日志写入文件地址
        ErrorOutputPaths: []string{"/tmp/_zap.log"},           // 系统error记录地址
        InitialFields: map[string]interface{}{ // 初始字段
            "app": "test",
        },
    }

    var err error
    Logger, err := config.Build()
    if err != nil {
        panic("zaplog init error:" + err.Error())
    }

    // 因对zap接口封装了一层方法，CallerSkip需加一
    Logger = Logger.WithOptions(zap.AddCallerSkip(1))
}

var (
    DebugLevel  = zapcore.DebugLevel
    InfoLevel   = zapcore.InfoLevel
    WarnLevel   = zapcore.WarnLevel
    ErrorLevel  = zapcore.ErrorLevel
    DPanicLevel = zapcore.DPanicLevel
    PanicLevel  = zapcore.PanicLevel
    FatalLevel  = zapcore.FatalLevel
)

type ZapConfig struct {
    LogPath    string        // 日志存放路径
    LogFile    string        // 日志文件基名
    MaxSize    int           // 日志最大兆字节数
    MaxAge     time.Duration // 日志最大存活天数
    FileLevel  zapcore.Level // 记录文件级别
    ConsLevel  zapcore.Level // 记录终端级别
    LocalTime  bool          // 是否以本地时间命名翻滚日志文件名
    Compress   bool          // 是否压缩旧日志
    CallerSkip int           // 针对zapcore原始接口封装了几层
}

func NewZapLogger(config ZapConfig) *zap.Logger {

    // rotating writer
    rollingConfig := RollingConfig{
        FileName:  path.Join(config.LogPath, config.LogFile),
        MaxAge:    config.MaxAge,
        MaxSize:   config.MaxSize,
        LocalTime: true,
        Compress:  false,
    }
    rollingWriter := zapcore.AddSync(RollingWriter(rollingConfig))
    consoleWriter := zapcore.AddSync(os.Stdout)

    // init encoder
    encoder := zap.NewDevelopmentEncoderConfig()
    encoder.EncodeTime = comfortableTimeEncoder

    // new _zap core
    fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoder), rollingWriter, zap.NewAtomicLevelAt(config.FileLevel))
    consCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoder), consoleWriter, zap.NewAtomicLevelAt(config.ConsLevel))

    options := []zap.Option{
        zap.AddCaller(),
        zap.AddCallerSkip(config.CallerSkip),
        zap.Development(),
        zap.ErrorOutput(nil),
        zap.AddStacktrace(FatalLevel),
    }

    // new logger
    return zap.New(zapcore.NewTee(fileCore, consCore), options...)
}

func GetLoggerLevel(lv string) zapcore.Level {
    switch lv {
    case "DEBUG":
        return DebugLevel
    case "INFO":
        return InfoLevel
    case "WARN":
        return WarnLevel
    case "ERROR":
        return ErrorLevel
    case "DPANIC":
        return DPanicLevel
    case "PANIC":
        return PanicLevel
    case "FATAL":
        return FatalLevel
    }
    return DebugLevel
}
