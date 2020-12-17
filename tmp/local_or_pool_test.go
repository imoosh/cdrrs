package tmp

import (
	uuid "github.com/satori/go.uuid"
	"sync"
	"testing"
)

const maxTimes = 100000
const maxRoutine = 36

func BenchmarkLocalOrPool(b *testing.B) {
	b.Run("pool", benchmarkPool)
	b.Run("local", benchmarkLocal)
}

func benchmarkPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for r := 0; r < maxRoutine; r++ {
			wg.Add(1)
			go func() {
				for i := 0; i < maxTimes; i++ {
					sip := New()
					sip.UserAgent = uuid.NewV4().String()
					sip.reset()
					Free(sip)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func benchmarkLocal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for r := 0; r < maxRoutine; r++ {
			wg.Add(1)
			go func() {
				for i := 0; i < maxTimes; i++ {
					sip := Local()
					sip.UserAgent = uuid.NewV4().String()
					sip.reset()
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
