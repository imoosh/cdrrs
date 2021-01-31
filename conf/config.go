package conf

import (
	"bytes"
	"centnet-cdrrs/common/cache/redis"
	"centnet-cdrrs/common/database/orm"
	"centnet-cdrrs/common/kafka"
	"centnet-cdrrs/common/log"
	xtime "centnet-cdrrs/common/time"
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
	FileParser *ParserConfig
}

type CDRConfig struct {
	CachedLife     xtime.Duration
	MaxFlushCap    int
	MaxCacheCap    int
	CdrFlushPeriod xtime.Duration
	CdrTablePeriod xtime.Duration
}

type ParserConfig struct {
	RootPath      string // 规则顶层路径
	StartDatePath string // 开始日期路径 202012/20201201
	MaxParseProcs int
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
