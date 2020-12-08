package conmap

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func BenchmarkNewConCurrentMap(b *testing.B) {
	var hm = NewExpireMap()

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

var hm = NewExpireMap()

func TestExpireMap_SetWithExpire(t *testing.T) {

	//hm.SetWithExpire("abc", 123, time.Second*3)
	//v, ok := hm.Get("abc")
	//t.Log(ok, v)
	//
	//time.Sleep(time.Second * 5)
	//v, ok = hm.Get("abc")
	//t.Log(ok, v)

	var wg sync.WaitGroup
	for a := 0; a < 100; a++ {
		wg.Add(1)
		go func(x int) {
			succ, fail := 0, 0
			for i := 0; i < 100000; i++ {
				k, v := fmt.Sprintf("%03d%06d", x, i), i*i
				hm.SetWithExpire(k, v, time.Second*3)
				//hm.Set(k, v)
				_, ok := hm.Get(k)
				if ok {
					succ++
				} else {
					fail++
				}
			}
			t.Log(succ, "--", fail)
			wg.Done()
		}(a)
	}
	wg.Wait()

	time.Sleep(time.Second * 10)

	for a := 0; a < 100; a++ {
		wg.Add(1)
		go func(x int) {
			succ, fail := 0, 0
			for i := 0; i < 100000; i++ {
				k, _ := fmt.Sprintf("%03d%06d", x, i), i*i
				//hm.SetWithExpire(k,v, time.Second*20)
				//hm.Set(k, v)
				_, ok := hm.Get(k)
				if ok {
					succ++
				} else {
					fail++
				}
			}
			t.Log(succ, "--", fail)
			wg.Done()
		}(a)
	}

	wg.Wait()
}

func BenchmarkExpireMap_SetWithExpire(b *testing.B) {
	var hm = NewExpireMap()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				for i := 0; i < 100000; i++ {
					hm.SetWithExpire(strconv.Itoa(i), i*i, time.Second*3)
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
	}
}
