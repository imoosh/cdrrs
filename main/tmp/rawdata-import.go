package main

import (
	"bufio"
	"fmt"
	"github.com/satori/go.uuid"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	n, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	for i := 0; i < n; i++ {
		now := time.Now().Format("20060102150405.000000")
		mock(os.Args[1], now)
	}
}

func mock(filepath, basename string) {
	var inviteMessage = `SIP/2.0 200 OK\0D\0AFrom: ""1101385""<sip:1101385@192.168.6.24;user=phone>;tag=04ab01e2d14221470\0D\0ATo: <sip:18113007510@219.143.187.139;user=phone>;tag=b0520e20-0-13e0-67e6bb-76fea845-67e6bb\0D\0ACall-ID: 04ab01e2d142787@192.168.6.24\0D\0A[Generated Call-ID: 04ab01e2d142787@192.168.6.24]\0D\0ACSeq: 17 INVITE\0D\0AAllow: INVITE,ACK,OPTIONS,REGISTER,INFO,BYE,UPDATE\0D\0AUser-Agent: DonJin SIP Server 3.2.0_i\0D\0ASupported: 100rel\0D\0AVia: SIP/2.0/UDP 192.168.6.24:5060;received=116.24.65.63;rport=5060;branch=z9hG4bK31596\0D\0AContact: <sip:018926798345@219.143.187.139;user=phone>\0D\0AContent-Type: application/sdp\0D\0AContent-Length: 126\0D\0A\0D\0Av=0\0D\0Ao=1101385 20000001 3 IN IP4 192.168.6.24\0D\0As=A call\0D\0Ac=IN IP4 192.168.6.24\0D\0At=0 0\0D\0Am=audio 10000 RTP/AVP 8 0\0D\0Aa=rtpmap:8 PCMA/8000\0D\0Aa=rtpmap:0 PCMU/8000\0D\0Aa=ptime:20\0D\0Aa=sendrecv\0D\0A`
	var rawInviteMessage = `"20201123120800","10020044201","220.248.118.20","9080","61.220.35.200","8080",` + "\"" + inviteMessage + "\""
	var byeMessage = `SIP/2.0 200 OK\0D\0AFrom: ""1101385""<sip:1101385@192.168.6.24;user=phone>;tag=04ab01e2d14221470\0D\0ATo: <sip:18113007510@219.143.187.139;user=phone>;tag=b0520e20-0-13e0-67e6bb-76fea845-67e6bb\0D\0ACall-ID: 04ab01e2d142787@192.168.6.24\0D\0A[Generated Call-ID: 04ab01e2d142787@192.168.6.24]\0D\0ACSeq: 18 BYE\0D\0AVia: SIP/2.0/UDP 192.168.6.24:5060;received=116.24.65.63;rport=5060;branch=z9hG4bK9701\0D\0ASupported: 100rel\0D\0AContent-Length: 0\0D\0A`
	var rawByeMessage = `"20201123120805","10020044201","220.248.118.20","9080","61.220.35.200","8080",` + "\"" + byeMessage + "\""

	const maxSize = 100000
	var uuidList [maxSize]string
	for i := 0; i < maxSize; i++ {
		uuidList[i] = uuid.NewV4().String()
	}

	fw1, err := os.Create(filepath + "/" + basename + ".invite.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fw1.Close()
	fmt.Printf("open %s success\n", filepath+"/"+basename+".invite.txt")
	bufWriter1 := bufio.NewWriterSize(fw1, 1024)

	//t := time.Now()
	for i := 0; i < maxSize; i++ {
		ss := strings.Split(rawInviteMessage, `\0D\0A`)
		for x, s := range ss {
			if strings.HasPrefix(s, "Call-ID: ") {
				ss[x] = "Call-ID: " + uuidList[i]
			}
		}
		rawInviteMessage = strings.Join(ss, `\0D\0A`)
		_, err := bufWriter1.WriteString(rawInviteMessage + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}
	err = bufWriter1.Flush()
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Printf("INVITE-200OK messages completed. %d packets, %v, %.f pps\n", maxSize, time.Since(t), maxSize/time.Since(t).Seconds())
	f, _ := os.Create(filepath + "/" + basename + ".invite.txt.ok")
	_ = f.Close()

	//time.Sleep(time.Second)

	//time.Sleep(time.Second * 5)
	//log.Debug("5 seconds sleeping...")

	fw2, err := os.Create(filepath + "/" + basename + ".bye.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fw2.Close()
	fmt.Printf("open %s success\n", filepath+"/"+basename+".bye.txt")
	bufWriter2 := bufio.NewWriterSize(fw2, 1024)

	//t = time.Now()
	for i := 0; i < maxSize; i++ {
		ss := strings.Split(rawByeMessage, `\0D\0A`)
		for x, s := range ss {
			if strings.HasPrefix(s, "Call-ID: ") {
				ss[x] = "Call-ID: " + uuidList[maxSize-i-1]
			}
		}
		rawByeMessage = strings.Join(ss, `\0D\0A`)
		_, _ = bufWriter2.WriteString(rawByeMessage + "\n")
	}
	_ = bufWriter2.Flush()
	//fmt.Printf("BYE-200OK messages completed. %d packets, %v, %.f pps\n", maxSize, time.Since(t), maxSize/time.Since(t).Seconds())
	f, _ = os.Create(filepath + "/" + basename + ".bye.txt.ok")
	_ = f.Close()
}
