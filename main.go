package main

import (
    "VoipSniffer/adapter/pcap"
    "VoipSniffer/dao"
    "VoipSniffer/prot/sip"
    "fmt"
    "github.com/astaxie/beego/orm"
    "sync"
)

func main() {
    pcap.StartCapture()
}

func main2() {

    sipPacket := dao.Sip{
        Sip:           "192.168.1.98",
        Sport:         5060,
        Dip:           "192.168.1.14",
        Dport:         5060,
        CallId:        "b5deab6380c4e57fa20486e493c68324",
        ReqMethod:     "INVITE",
        ReqStatusCode: 5060,
        ReqUser:       "wayne",
        ReqHost:       "dvao.cn",
        ReqPort:       5060,
        FromName:      "wayne",
        FromUser:      "wayne",
        FromHost:      "dvao.cn",
        FromPort:      5060,
        ToName:        "abc",
        ToUser:        "abc",
        ToHost:        "asdkl;jf.ca",
        ToPort:        5060,
        ContactName:   "abc",
        ContactUser:   "abc",
        ContactHost:   "ak;lfa.cn",
        ContactPort:   5060,
        CseqMethod:    "INVITE",
        UserAgent:     "cnx3000",
    }

    o := orm.NewOrm()
    n, err := o.Insert(&sipPacket)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(n)
}

func main1() {
    // Load up a test message
    raw := []byte("SIP/2.0 200 OK\r\n" +
        "Via: SIP/2.0/UDP 192.168.2.242:5060;received=22.23.24.25;branch=z9hG4bK5ea22bdd74d079b9;alias;rport=5060\r\n" +
        "To: <sip:JohnSmith@mycompany.com>;tag=aprqu3hicnhaiag03-2s7kdq2000ob4\r\n" +
        "From: sip:HarryJones@mycompany.com;tag=89ddf2f1700666f272fb861443003888\r\n" +
        "CSeq: 57413 REGISTER\r\n" +
        "Call-ID: b5deab6380c4e57fa20486e493c68324\r\n" +
        "Contact: <sip:JohnSmith@192.168.2.242:5060>;expires=192\r\n\r\n")

    var wg sync.WaitGroup
    wg.Add(6)

    for a := 0; a < 6; a++ {
        go func() {
            for i := 0; i < 1; i++ {
                sip.Parse(raw)
                sip := sip.Parse(raw)
                fmt.Println("From: ", string(sip.From.User), " To: ", string(sip.To.User))
                fmt.Println(string(sip.Contact.User), string(sip.Req.User))
                fmt.Println(string(sip.Req.StatusCode), string(sip.Req.Method), string(sip.Cseq.Method))
            }
            wg.Done()
        }()
    }

    wg.Wait()
}
