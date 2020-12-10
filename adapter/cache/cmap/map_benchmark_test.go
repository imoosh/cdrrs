package concurrent_map

import (
    "strconv"
    "sync"
    "testing"
    "time"
)

type Map interface {
    Set(key interface{}, val interface{})
    Get(key interface{}) (interface{}, bool)
    Del(key interface{})
}

func benchmarkMap(b *testing.B, hm Map) {
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                for i := 0; i < 100000; i++ {
                    hm.Set(strconv.Itoa(i), i*i)
                    hm.Set(strconv.Itoa(i), i*i)
                    hm.Del(strconv.Itoa(i))
                }
                wg.Done()
            }()
        }
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                for i := 0; i < 100000; i++ {
                    hm.Get(strconv.Itoa(i))
                }
                wg.Done()
            }()
        }
        wg.Wait()
    }
}

func BenchmarkSyncmap(b *testing.B) {
    b.Run("map with RWLock", func(b *testing.B) {
        hm := CreateRWLockMap()
        benchmarkMap(b, hm)
    })

    b.Run("sync.map", func(b *testing.B) {
        hm := CreateSyncMapBenchmarkAdapter()
        benchmarkMap(b, hm)
    })

    b.Run("concurrent map", func(b *testing.B) {
        superman := CreateConcurrentMapBenchmarkAdapter(199)
        benchmarkMap(b, superman)
    })
}

func BenchmarkConcurrentMap_Expire(b *testing.B) {
    hm := CreateConcurrentMap(99, 20*time.Second)
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                for i := 0; i < 100000; i++ {
                    hm.SetEx(StrKey(strconv.Itoa(i)), i*i)
                }
                wg.Done()
            }()
        }
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                for i := 0; i < 100000; i++ {
                    hm.Get(StrKey(strconv.Itoa(i)))
                }
                wg.Done()
            }()
        }
        wg.Wait()
    }
}
