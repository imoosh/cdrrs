package tmp

import (
	"github.com/RussellLuo/timingwheel"
	"sync"
	"time"
)

type MemCache struct {
	cache map[string]string
	lock  sync.RWMutex
}

var mc *MemCache
var tw *timingwheel.TimingWheel

func Init() {
	tw = timingwheel.NewTimingWheel(time.Second, 300)
	tw.Start()
	//defer tw.Stop()

	mc = &MemCache{
		cache: make(map[string]string),
	}
}
