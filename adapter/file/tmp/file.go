package tmp

import (
	"bufio"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/model"
	"centnet-cdrrs/model/prot/sip"
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var cdrLock sync.Mutex

//
//type DataTransfer struct {
//	TopFilePath     string
//	StartMonthPath   string
//	CurrentDaysPath string
//}
//
//func NewFileReader(filepath string) (*DataTransfer, error) {
//	_, err := os.Stat(filepath)
//	if err != nil {
//		return nil, err
//	}
//
//	file, err := os.OpenFile(filepath, os.O_RDONLY, 0)
//	if err != nil {
//		return nil, err
//	}
//
//	return &DataTransfer{
//		FilePath: filepath,
//		File:     file,
//	}, nil
//}
//
//func (df *DataTransfer) Read(b []byte) (int, error) {
//	return df.File.Read(b)
//}
//
//func (df *DataTransfer) Close() error {
//	return df.File.Close()
//}
//
//func (df *DataTransfer) FlushFileList() {
//	ioutil.ReadDir(df.TopFilePath+"/"+df.CurrentDaysPath)
//}
//
//func getFileCTime(file string) int64 {
//	osType := runtime.GOOS
//
//	fileInfo, _ := os.Stat(file)
//	if osType == "linux" {
//		stat := fileInfo.Sys().(*syscall.Stat_t)
//		dtime := int64(stat.Ctim.Sec)
//		return dtime
//	} else if osType == "darwin" {
//		//stat := fileInfo.Sys().(*syscall.Stat_t)
//		//dtime := stat.Ctimespec
//		//return dtime
//	}
//	return time.Now().Unix()
//}
//
//func (df *DataTransfer) Run() {
//
//}

func atoi(s string, n int) (int, error) {
	if len(s) == 0 {
		return n, nil
	}

	return strconv.Atoi(s)
}

func Run(InvitePath, bytePath, output string) error {
	fw, err := os.Create(output + "/" + "cdr.txt")
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debugf("open %s success", output+"/"+"cdr.txt")
	bufWriter := bufio.NewWriterSize(fw, 10<<20)

	fileInfo, err := ioutil.ReadDir(bytePath)
	if err != nil {
		log.Error(err)
		return errors.New("ioutil.ReadDir Error")
	}

	var wg sync.WaitGroup
	for _, info := range fileInfo {
		if !strings.HasSuffix(info.Name(), ".txt") || strings.HasSuffix(info.Name(), ".sip.txt") {
			continue
		}

		wg.Add(1)
		go func(f string) {
			defer wg.Done()

			fr, err := os.Open(bytePath + "/" + f)
			if err != nil {
				log.Error(err)
				return
			}
			defer fr.Close()
			log.Debugf("open %s success", bytePath+"/"+f)

			bufReader := bufio.NewReaderSize(fr, 10<<20)

			for {
				line, _, err := bufReader.ReadLine()
				if err != nil && err != io.EOF {
					panic(err)
				} else if err == io.EOF {
					break
				}

				parseLine(bufWriter, string(line))
			}
		}(info.Name())
	}

	wg.Wait()

	fileInfo, err = ioutil.ReadDir(InvitePath)
	if err != nil {
		log.Error(err)
		return errors.New("ioutil.ReadDir Error")
	}

	for _, info := range fileInfo {
		if !strings.HasSuffix(info.Name(), ".txt") || strings.HasSuffix(info.Name(), ".sip.txt") {
			continue
		}

		wg.Add(1)
		go func(f string) {
			defer wg.Done()

			fr, err := os.Open(InvitePath + "/" + f)
			if err != nil {
				log.Error(err)
				return
			}
			defer fr.Close()
			log.Debugf("open %s success", InvitePath+"/"+f)

			bufReader := bufio.NewReaderSize(fr, 10<<20)

			for {
				line, _, err := bufReader.ReadLine()
				if err != nil && err != io.EOF {
					panic(err)
				} else if err == io.EOF {
					break
				}

				parseLine(bufWriter, string(line))
			}
		}(info.Name())
	}

	wg.Wait()
	log.Debugf("%d CDRs written", count)

	bufWriter.Flush()
	fw.Close()
	return nil
}

var count int

func parseLine(writer *bufio.Writer, line string) {
	var err error
	rtd, err := ParseLine(line)
	if err != nil {
		return
	}
	sipMsg := sip.Parse([]byte(rtd.ParamContent))

	//sip.PrintSipStruct(&sipMsg)

	pkt := model.AnalyticSipPacket{
		EventId:       rtd.EventId,
		EventTime:     rtd.EventTime,
		Sip:           rtd.SaddrV4,
		Sport:         0,
		Dip:           rtd.DaddrV4,
		Dport:         0,
		CallId:        string(sipMsg.CallId.Value),
		CseqMethod:    string(sipMsg.Cseq.Method),
		ReqMethod:     string(sipMsg.Req.Method),
		ReqStatusCode: 0,
		ReqUser:       string(sipMsg.Req.User),
		ReqHost:       string(sipMsg.Req.Host),
		ReqPort:       0,
		FromName:      string(sipMsg.From.Name),
		FromUser:      string(sipMsg.From.User),
		FromHost:      string(sipMsg.From.Host),
		FromPort:      0,
		ToName:        string(sipMsg.To.Name),
		ToUser:        string(sipMsg.To.User),
		ToHost:        string(sipMsg.To.Host),
		ToPort:        0,
		ContactName:   string(sipMsg.Contact.Name),
		ContactUser:   string(sipMsg.Contact.User),
		ContactHost:   string(sipMsg.Contact.Host),
		ContactPort:   0,
		UserAgent:     string(sipMsg.Ua.Value),
	}

	// 没有call-id、cseq.method、直接丢弃
	if len(pkt.CallId) == 0 || len(pkt.CseqMethod) == 0 || len(pkt.ToUser) == 0 {
		return
	}

	// 被叫号码字段未解析出手机号码或坐席号码归属地，直接丢弃(同一会话中所有包的FROM字段或TO字段都一样)
	//pkt.CalleeInfo, err = model.parseCalleeInfo(pkt.ToUser)
	//if err != nil {
	//	//log.Debug("DIRTY-DATA:", string(value.([]byte)))
	//	return
	//}

	if pkt.Sport, err = atoi(rtd.Sport, 0); err != nil {
		return
	}
	if pkt.Dport, err = atoi(rtd.Dport, 0); err != nil {
		return
	}
	if pkt.ReqStatusCode, err = atoi(string(sipMsg.Req.StatusCode), 0); err != nil {
		return
	}
	if pkt.ReqPort, err = atoi(string(sipMsg.Req.Port), 5060); err != nil {
		return
	}
	if pkt.FromPort, err = atoi(string(sipMsg.From.Port), 5060); err != nil {
		return
	}
	if pkt.ToPort, err = atoi(string(sipMsg.To.Port), 5060); err != nil {
		return
	}
	if pkt.ContactPort, err = atoi(string(sipMsg.Contact.Port), 5060); err != nil {
		return
	}

	jsonStr, err := json.Marshal(pkt)
	if err != nil {
		log.Error(err)
		return
	}

	if pkt.CseqMethod == "INVITE" && pkt.ReqStatusCode == 200 {
		cdr := HandleInvite200OKMessage(pkt.CallId, pkt)
		if len(cdr) != 0 {
			cdrLock.Lock()
			defer cdrLock.Unlock()
			_, err := writer.WriteString(cdr + "\n")
			if err != nil {
				log.Error(err)
			}
			count++
		}
	} else if pkt.CseqMethod == "BYE" && pkt.ReqStatusCode == 200 {
		HandleBye200OKMsg(pkt.CallId, string(jsonStr))
	}
}

func HandleInvite200OKMessage(k string, pkt model.AnalyticSipPacket) string {
	mc.lock.RLock()
	value, ok := mc.cache[k]
	if !ok {
		mc.lock.RUnlock()
		return ""
	}
	mc.lock.RUnlock()

	var bye200OKMsg model.AnalyticSipPacket
	if err := json.Unmarshal([]byte(value), &bye200OKMsg); err != nil {
		return ""
	}

	connectTime, err := time.ParseInLocation("20060102150405", pkt.EventTime, time.Local)
	if err != nil {
		log.Errorf("time.ParseLine error: %s", pkt.EventTime)
		return ""
	}
	disconnectTime, err := time.ParseInLocation("20060102150405", bye200OKMsg.EventTime, time.Local)
	if err != nil {
		log.Errorf("time.ParseLine error: %s", bye200OKMsg.EventTime)
		return ""
	}

	//填充话单字段信息
	now := time.Now()
	cdr := dao.VoipRestoredCdr{
		CallId:         pkt.CallId,
		Uuid:           uuid.NewV4().String(),
		CallerIp:       pkt.Dip,
		CallerPort:     pkt.Dport,
		CalleeIp:       pkt.Sip,
		CalleePort:     pkt.Sport,
		CallerNum:      pkt.FromUser,
		CalleeNum:      pkt.ToUser,
		CalleeDevice:   pkt.UserAgent,
		CalleeProvince: "",
		CalleeCity:     "",
		ConnectTime:    connectTime.Unix(),
		DisconnectTime: disconnectTime.Unix(),
		Duration:       0,
		FraudType:      "",
		CreateTime:     now.Format("2006-01-02 15:04:05"),
		CreateTimeX:    now,
	}

	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	if pkt.Dip == bye200OKMsg.Sip {
		cdr.CallerDevice = bye200OKMsg.UserAgent
	} else {
		//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
		cdr.CallerDevice = ""
	}
	cdr.Duration = int(disconnectTime.Sub(connectTime).Seconds())

	// 填充原始被叫号码及归属地
	//calleeInfo := bye200OKMsg.CalleeInfo
	//cdr.CalleeNum = calleeInfo.Number
	//cdr.CalleeProvince = calleeInfo.Pos.Province
	//cdr.CalleeCity = calleeInfo.Pos.City

	// 推送至诈骗分析模型
	cdrStr, err := json.Marshal(&cdr)
	if err != nil {
		log.Errorf("json.Marshal error: ", err)
		return ""
	}

	// 读到即删除
	mc.lock.Lock()
	defer mc.lock.Unlock()
	delete(mc.cache, k)

	return string(cdrStr)
}

func HandleBye200OKMsg(k, v string) {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	mc.cache[k] = v

	tw.AfterFunc(time.Minute*30, func() {
		mc.lock.Lock()
		defer mc.lock.Unlock()
		delete(mc.cache, k)
	})
}
