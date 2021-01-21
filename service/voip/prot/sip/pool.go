package sip

import "sync"

var sipPool = sync.Pool{New: func() interface{} { return new(SipMsg) }}
