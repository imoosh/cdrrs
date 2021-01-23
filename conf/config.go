package conf

import (
    "bytes"
    "centnet-cdrrs/common/cache/redis"
    "centnet-cdrrs/common/database/orm"
    "centnet-cdrrs/common/kafka"
    "centnet-cdrrs/common/log"
    xtime "centnet-cdrrs/common/time"
    "centnet-cdrrs/service/adapters/file"
    "fmt"
    "github.com/BurntSushi/toml"
)

var (
    Conf = &Config{}
)

type Config struct {
    Logging    *log.Config
    Kafka      *kafka.Config
    ORM        *orm.Config
    CDR        *CDRConfig
    Redis      *redis.Config
    FileParser *file.Config
}

type CDRConfig struct {
    CachedLife     xtime.Duration
    MaxFlushCap    int
    MaxCacheCap    int
    CdrFlushPeriod xtime.Duration
    CdrTablePeriod xtime.Duration
}

func (c *Config) String() string {
    b := &bytes.Buffer{}
    err := toml.NewEncoder(b).Encode(c)
    if err != nil {
        return ""
    }

    return b.String()
}

func Init() {
    if len(configFilePath) == 0 {
        panic("parse config file error")
    }

    _, err := toml.DecodeFile(configFilePath, Conf)
    if err != nil {
        panic(fmt.Sprintf("toml.DecodeFile: %v\n", err))
    }
}
