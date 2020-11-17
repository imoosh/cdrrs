package file

import (
    "fmt"
    "os"
    "strings"
)

const SEPARATOR = "\r\n"


//数据格式: "event_time", "event_id", saddr_v4", "daddr_v4", "dport", "param_content"，每条数据以"\r\n"分割

type RawTextData struct {
    EventTime    string
    EventId      string
    SaddrV4      string
    Sport        string
    DaddrV4      string
    Dport        string
    ParamContent string
}

func Pretreatment(s string) string {
    s = strings.ReplaceAll(s, "\\0D\\0A", "\r\n")
    return strings.ReplaceAll(s, "\"\"", "\"")
}

func Split(s string) (error, []string) {
    // 分割成7端

    ss := strings.SplitN(s, "\",\"", 7)
    if len(ss) != 7 {
        return fmt.Errorf("split error: %s", s), nil
    }

    for _, val := range ss {
        if len(val) == 0 {
            return fmt.Errorf("split error: %s", s), nil
        }
    }

    ss[0] = ss[0][1:]
    ss[6] = ss[6][:len(ss[6])-1]

    return nil, ss
}

func Parse(ss []string) (rtd RawTextData) {
    rtd.EventTime = ss[0]
    rtd.EventId = ss[1]
    rtd.SaddrV4 = ss[2]
    rtd.Sport = ss[3]
    rtd.DaddrV4 = ss[4]
    rtd.Dport = ss[5]
    rtd.ParamContent = ss[6]

    return
}
