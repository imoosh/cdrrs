package conf

import "flag"

var (
	configFilePath string
	//InvitePath     string
	//ByePath        string
	//DemoOutputPath string
)

func init() {
	//flag.StringVar(&configFilePath, "c", "conf/config.toml", "default configuration path")
	flag.StringVar(&configFilePath, "c", "../../scripts/config.toml", "default configuration path")
	//flag.StringVar(&InvitePath, "i", ".", "invite file path")
	//flag.StringVar(&ByePath, "b", ".", "bye file path")
	//flag.StringVar(&DemoOutputPath, "o", ".", "output file path")
}
