package conf

import (
	"flag"
)

var (
	configFilePath string
)

func init() {
	flag.StringVar(&configFilePath, "c", "conf/config.toml", "default configuration path")
}
