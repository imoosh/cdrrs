package main

import (
    "centnet-cdrrs/common/log"
    "centnet-cdrrs/conf"
    "centnet-cdrrs/service"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "runtime"
    "runtime/debug"
    "syscall"
    "time"
)

var (
    showVersion bool
    BuiltID     string
    BuiltHost   string
    BuiltTime   string
    GoVersion   string
)

func init() {
    flag.BoolVar(&showVersion, "v", false, "show application version and exit")

    if !flag.Parsed() {
        flag.Parse()
    }

    if showVersion {
        fmt.Println(getAppVersion())
        os.Exit(0)
    }
}

func getAppVersion() string {
    return fmt.Sprintf(""+
        "Built ID:   %s\n"+
        "Built Host: %s\n"+
        "Built Time: %s\n"+
        "Go Vesrion: %s\n",
        BuiltID, BuiltHost, BuiltTime, GoVersion)
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    /* 解析参数 */
    flag.Parse()
    conf.Init()
    fmt.Println(conf.Conf)

    /* 日志模块初始化 */
    log.Init(conf.Conf.Logging)

    service.New(conf.Conf)

    printMemStatsLoop()

    go func() {
        time.Sleep(time.Minute * 12)
        log.Debug("debug.SetGCPercent(10) before")
        printMemStats()

        log.Debug("debug.SetGCPercent(10)")
        debug.SetGCPercent(10)
        time.Sleep(time.Minute)

        log.Debug("debug.SetGCPercent(10) after")
        printMemStats()
    }()

    // os signal
    sigterm := make(chan os.Signal, 1)
    signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
    <-sigterm
}

func printMemStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    log.Debugf("Alloc = %vMB Sys = %vMB NumGC = %v",
        m.Alloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

func printMemStatsLoop() {
    go func() {
        for {
            time.Sleep(time.Minute)
            printMemStats()
        }
    }()
}
