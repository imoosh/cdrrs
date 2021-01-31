package service

import (
	"centnet-cdrrs/common/cache/local"
	"centnet-cdrrs/common/kafka"
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/model"
	"encoding/json"
	"fmt"
	"time"
)

var (
	_cdrCh = make(chan *model.VoipCDR)
)

type Service struct {
	c          *conf.Config
	parser     *Parser
	dao        *dao.Dao
	fraudModel *kafka.Producer
	mc         *local.ShardedCache
	synth      *Synth
	cdrs       []*model.VoipCDR
	_cdrCh     chan *model.VoipCDR
}

func New(c *conf.Config) (svr *Service, err error) {

	// kafka交互模块
	var fm *kafka.Producer
	if c.Kafka.FraudModel.Enable {
		if fm, err = kafka.NewProducer(c.Kafka.FraudModel); err != nil {
			log.Fatal(err)
		}
		fm.Run()
	}

	// 预读号码归属地缓存
	dao := dao.New(c)
	model.CacheNumberPositions(dao.CacheNumberPositions())

	// 初始化数据解析器
	svr = &Service{
		c:          c,
		parser:     NewParser(c.FileParser),
		dao:        dao,
		fraudModel: fm,
	}

	synth := newSynthesizer(svr)
	synth.OnExpired(func(args interface{}) {
		items := args.(map[string]*model.SipCache)
		for k, v := range items {
			if v != nil {
				if c := model.NewExpiredCDR(k, v); c != nil {
					Collect(c)
				}
				v.Free()
			}
		}
	})
	synth.Run()
	svr.synth = synth

	// 启动话单处理服务
	go svr.handleCDR()

	// 最后启动数据解析器
	svr.parser.SetParseFunc(svr.DoLine)
	svr.parser.Run()

	return svr, nil
}

func (srv *Service) DoLine(line interface{}) {

	// 解析sip报文
	var sip model.SipPacket
	if nil != model.ParseSipPacket(line, &sip) {
		return
	}

	// invite-200ok消息
	if sip.CseqMethod == "INVITE" && sip.ReqStatusCode == 200 {
		invItem := model.NewSipItemFromSipPacket(&sip)
		invItem.Type = model.SipStatusInvite200OK
		invItem.CalleeDevice = sip.UserAgent
		invItem.ConnectTime = sip.EventTime
		srv.synth.Input(invItem)
	} else if sip.CseqMethod == "BYE" && sip.ReqStatusCode == 200 {
		byeItem := model.NewSipItemFromSipPacket(&sip)
		byeItem.Type = model.SipStatusBye200OK
		byeItem.DisconnectTime = sip.EventTime
		srv.synth.Input(byeItem)
	} else {
		log.Debug("no handler for else condition")
	}
}

func Collect(vc interface{}) {
	_cdrCh <- vc.(*model.VoipCDR)
}

func (srv *Service) handleCDR() {
	var (
		duration     time.Duration
		procDBTime   time.Time
		curTableName string
	)

	duration = time.Duration(srv.c.CDR.CdrFlushPeriod)
	ticker := time.NewTicker(duration)

	for {
		select {
		case cdr := <-_cdrCh:
			// 在一个go routine里处理所有话单，确保时间id是顺序产生的
			now := time.Now()
			cdr.SetCDRId(now)
			cdr.SetCreateTime(now)
			cdr.SetCDRTable(now, time.Duration(srv.c.CDR.CdrTablePeriod))

			srv.pushCDRToKafka(cdr)

			if cdr.TableName != curTableName {
				srv.dao.CreateTable(cdr.TableName)
				// 不是第一次建表，将缓存话单先入旧库表
				if len(curTableName) != 0 {
					srv.flushCDRToDB(curTableName, srv.cdrs)
					procDBTime = time.Now()
				}
				curTableName = cdr.TableName
			}

			srv.cdrs = append(srv.cdrs, cdr)
			if len(srv.cdrs) == srv.c.CDR.MaxFlushCap {
				srv.flushCDRToDB(curTableName, srv.cdrs)
				procDBTime = time.Now()
			}
		case <-ticker.C:
			// 有一段时间没刷新数据库
			now := time.Now()
			if len(srv.cdrs) != 0 && now.Sub(procDBTime) >= duration {
				srv.flushCDRToDB(curTableName, srv.cdrs)
				procDBTime = now
			}
		}
	}
}

func (srv *Service) appendCDRs(c *model.VoipCDR) {
	srv.cdrs = append(srv.cdrs, c)
}

func (srv *Service) clearCDRs() {
	for _, x := range srv.cdrs {
		x.Free()
	}
	srv.cdrs = srv.cdrs[:0]
}

func (srv *Service) pushCDRToKafka(c *model.VoipCDR) {
	// 推送至诈骗分析模型
	if srv.c.Kafka.FraudModel.Enable {
		cdrStr, err := json.Marshal(c)
		if err != nil {
			log.Errorf("json.Marshal error: ", err)
			return
		}

		srv.fraudModel.Log(fmt.Sprintf("%d", c.Id), string(cdrStr))
	}
}

func (srv *Service) flushCDRToDB(tblName string, cdrs []*model.VoipCDR) {
	if len(cdrs) == 0 {
		return
	}
	srv.dao.MultiInsertCDR(tblName, cdrs)
	srv.clearCDRs()
}
