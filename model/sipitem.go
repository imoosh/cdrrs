package model

import "sync"

var (
	emptySipItem = SipItem{}
	sipItemPool  = sync.Pool{New: func() interface{} { return &SipItem{} }}
)

type SipItem struct {
	CallId         string
	Caller         string
	Callee         string
	SrcIP          string
	DestIP         string
	SrcPort        uint16
	DestPort       uint16
	CallerDevice   string
	CalleeDevice   string
	ConnectTime    string
	DisconnectTime string
}

func NewSipItem() *SipItem {
	return sipItemPool.Get().(*SipItem)
}

func (pkt *SipItem) Free() {
	*pkt = emptySipItem
	sipItemPool.Put(pkt)
}
