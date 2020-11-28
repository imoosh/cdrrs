package dao

import (
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestMultiInsertCDR(t *testing.T) {
	cdr := VoipRestoredCdr{
		Id:             0,
		CallId:         "1234",
		Uuid:           uuid.NewV4().String(),
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
		ConnectTime:    0,
		DisconnectTime: 0,
		Duration:       0,
		FraudType:      "",
		CreateTime:     "",
	}

	cdrs := []*VoipRestoredCdr{
		&cdr,
	}
	MultiInsertCDR(cdrs)
}
