package orm

import (
    "centnet-cdrrs/common/log"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "time"
)

type Config struct {
    DSN         string        // database source name
    Active      int           // pool
    Idle        int           // pool
    IdleTimeout time.Duration // connect max life time
}

func NewMySQL(c *Config) (db *gorm.DB) {
    ormDB, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{})
    if err != nil {
        log.Error(err)
        panic(err)
    }

    // 设置日志
    ormDB.Config.Logger = logger.New(&ormLog{}, logger.Config{
        SlowThreshold: 200 * time.Millisecond,
        LogLevel:      logger.Error,
        Colorful:      false,
    })

    sqlDB, err := ormDB.DB()
    if err != nil {
        log.Error(err)
        panic(err)
    }

    sqlDB.SetMaxIdleConns(c.Idle)
    sqlDB.SetMaxOpenConns(c.Active)
    sqlDB.SetConnMaxLifetime(c.IdleTimeout)

    return ormDB
}

type ormLog struct {
}

func (l ormLog) Printf(format string, v ...interface{}) {
    log.Warnf(format, v...)
}
