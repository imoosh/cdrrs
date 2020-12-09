package conf

import (
	"bytes"
	"centnet-cdrrs/adapter/file"
	"centnet-cdrrs/adapter/kafka"
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"fmt"
	"github.com/BurntSushi/toml"
)

var (
	Conf = &Config{}
)

type Config struct {
	Logging    *log.Config
	Kafka      *kafka.Config
	Mysql      *dao.Config
	Redis      *redis.Config
	FileParser *file.Config
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
