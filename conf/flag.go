package conf

import "flag"

var (
	configFilePath string
)

func init() {
	//flag.StringVar(&configFilePath, "c", "conf/config.toml", "default configuration path")
	flag.StringVar(&configFilePath, "c", "../../scripts/config.toml", "default configuration path")

}
