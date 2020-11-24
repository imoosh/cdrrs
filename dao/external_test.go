package dao

import (
	"testing"
	"time"
)

func TestMultiInsertCDR(t *testing.T) {
	cdr := VoipRestoredCdr{
		Id:             0,
		CallId:         "1234",
		CallerIp:       "aaa",
		CallerPort:     0,
		CalleeIp:       "aaa",
		CalleePort:     0,
		CallerNum:      "aaa",
		CalleeNum:      "aa",
		CallerDevice:   "df",
		CalleeDevice:   "dfasd",
		CalleeProvince: "asdf",
		CalleeCity:     "asdf",
		ConnectTime:    time.Time{},
		DisconnectTime: time.Time{},
		Duration:       0,
		FraudType:      "",
	}

	cdrs := []*VoipRestoredCdr{
		&cdr,
	}
	MultiInsertCDR(cdrs)
}
