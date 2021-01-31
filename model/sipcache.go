package model

import (
	"sync"
)

var (
	emptySipItem = SipCache{}
	sipItemPool  = sync.Pool{New: func() interface{} { return &SipCache{} }}
)

const (
	SipRequestInvite = iota
	SipRequestBye
	SipStatusInvite200OK
	SipStatusBye200OK
)

type SipCache struct {
	Type           int    `json:"t"`
	CallId         string `json:"-"`
	Caller         string `json:"cr"`
	Callee         string `json:"ce"`
	SrcIP          string `json:"si"`
	DestIP         string `json:"di"`
	SrcPort        uint16 `json:"sp"`
	DestPort       uint16 `json:"dp"`
	CallerDevice   string `json:"cd"`
	CalleeDevice   string `json:"ed"`
	ConnectTime    string `json:"ct"`
	DisconnectTime string `json:"dt"`
}

func NewSipItem() *SipCache {
	return sipItemPool.Get().(*SipCache).Clear()
}

func (pkt *SipCache) Clear() *SipCache {
	*pkt = emptySipItem
	return pkt
}

func (pkt *SipCache) Free() {
	sipItemPool.Put(pkt)
}

func NewSipItemFromSipPacket(sip *SipPacket) *SipCache {
	item := NewSipItem()
	item.CallId = sip.CallId
	item.Caller = sip.FromUser
	item.Callee = sip.ToUser
	item.SrcIP = sip.Sip
	item.DestIP = sip.Dip
	item.SrcPort = uint16(sip.Sport)
	item.DestPort = uint16(sip.Dport)
	return item
}
