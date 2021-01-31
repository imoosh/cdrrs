package main

import (
    "fmt"
    "time"
)

func main() {
    var _setIdsTimeFormat = "20060102150405.000000"
    fmt.Println(time.Now().Format(_setIdsTimeFormat))
}
