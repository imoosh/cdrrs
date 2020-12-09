package model

type Config struct {

	// 输出话单级别 1: 被叫为手机号码，2：被叫为手机或座机，3：被叫为所有有效号码
	CDRRestoreLevelLevel int `json:"cdrRestoreLevel"`
}
