package main

import "fmt"

func main() {
	//t := time.Now()
	//ns := time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	//time.ParseInLocation("20060102150405", t.EventTime, time.Local)
	//ns, _ := time.Parse("15:04:05.000000000", time.Now().Format("15:04:05.000000000"))

	var args interface{}
	todo := make(chan interface{}, 1)

	go func() {
		for {
			select {
			case a := <-todo:
				fmt.Println(a)
			default:
			}
		}

	}()

	todo <- args
	todo <- args
	todo <- args

	select {}
	//var sn atomic.Int64
	//
	//now := time.Now()
	//id := int64((now.Hour()*60*60+now.Minute()*60+now.Second())*1e9+now.Nanosecond()/1e3) + sn.Add(1)%1e3
	//fmt.Println(now.Hour(), now.Minute(), now.Second(), id)
}
