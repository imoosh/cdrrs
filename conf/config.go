package conf

import (
    "bytes"
    "centnet-cdrrs/common/database/orm"
    "centnet-cdrrs/common/kafka"
    "centnet-cdrrs/common/log"
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
    FileParser *file.Config
}

type CDRConfig struct {
    CachedLife     int
    MaxFlushCap    int
    MaxCacheCap    int
    CdrFlushPeriod int
    CdrTablePeriod int
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
