package tmp

import (
	"github.com/RussellLuo/timingwheel"
	"sync"
)

type MemCache struct {
	cache map[string]string
	lock  sync.RWMutex
}

var mc *MemCache
var tw *timingwheel.TimingWheel

func Init() {
	mc = &MemCache{
		cache: make(map[string]string),
	}
}
