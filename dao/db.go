package dao

import (
	"centnet-cdrrs/library/log"
	"errors"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

var asyncDao AsyncDao

var fixedPhoneNumberAttributionMap = make(map[string]interface{})

var mobilePhoneNumberAttributionMap = make(map[string]interface{})

type Config struct {
	DSN              string
	FlushPeriod      int
	MaxFlushCapacity int
	MaxCacheCapacity int
}

type AsyncDao struct {
	msgQ      chan *VoipRestoredCdr
	wg        sync.WaitGroup
	closeChan chan struct{}
	c         *Config
}

func (ad *AsyncDao) Run() {
	var cache []*VoipRestoredCdr
	duration := time.Duration(ad.c.FlushPeriod) * time.Second
	timer := time.NewTimer(duration)

	go func() {
		for {
			select {
			case m := <-ad.msgQ:
				cache = append(cache, m)
				if len(cache) == ad.c.MaxFlushCapacity {
					MultiInsertCDR(cache)
					cache = cache[:0]
				}
			case <-timer.C:
				MultiInsertCDR(cache)
				cache = cache[:0]
				timer.Reset(duration)
			}
		}
	}()
}

func (ad *AsyncDao) LogCDR(cdr *VoipRestoredCdr) {
	ad.msgQ <- cdr
}

type VoipRestoredCdr struct {
	Id             int64  `json:"id"`
	Uuid           string `json:"uuid"`
	CallId         string `json:"callId"`
	CallerIp       string `json:"callerIp"`
	CallerPort     int    `json:"callerPort"`
	CalleeIp       string `json:"calleeIp"`
	CalleePort     int    `json:"calleePort"`
	CallerNum      string `json:"callerNum"`
	CalleeNum      string `json:"calleeNum"`
	CallerDevice   string `json:"callerDevice"`
	CalleeDevice   string `json:"calleeDevice"`
	CalleeProvince string `json:"calleeProvince"`
	CalleeCity     string `json:"calleeCity"`
	ConnectTime    int64  `json:"connectTime"`
	DisconnectTime int64  `json:"disconnectTime"`
	Duration       int    `json:"duration"`
	FraudType      string `json:"fraudType"`
	CreateTime     string `json:"createTime"`
}

// json序列化只需要省市字段
type PhonePosition struct {
	Id         int    `json:"-"`
	Prefix     string `json:"-"`
	Phone      string `json:"-"`
	Province   string `json:"province"`
	ProvinceId int64  `json:"-"`
	City       string `json:"city"`
	CityId     int64  `json:"-"`
	Isp        string `json:"-"`
	Code1      string `json:"-"`
	Zip        string `json:"-"`
	Types      string `json:"-"`
}

func Init(c *Config) error {
	//orm.Debug = true
	err := orm.RegisterDataBase("default", "mysql", c.DSN, 30)
	if err != nil {
		log.Error(err)
		return errors.New("orm.RegisterDataBase failed")
	}

	orm.RegisterModel(new(VoipRestoredCdr))
	orm.RegisterModel(new(PhonePosition))

	// 将号码归属地表预读到内存中，加速查询速度
	if err := CachePhoneNumberAttribution(); err != nil {
		log.Error(err)
		return errors.New("CachePhoneNumberAttribution failed")
	}

	asyncDao = AsyncDao{
		msgQ:      make(chan *VoipRestoredCdr, c.MaxCacheCapacity),
		closeChan: make(chan struct{}),
		wg:        sync.WaitGroup{},
		c:         c,
	}
	asyncDao.Run()

	return nil
}

//type DateTime time.Time
//
//func (t DateTime) MarshalJSON() ([]byte, error) {
//    var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
//    return []byte(stamp), nil
//}
//
//func (cdr VoipRestoredCdr) MarshalJSON() ([]byte, error) {
//    type TmpJSON VoipRestoredCdr
//    return json.Marshal(&struct {
//        TmpJSON
//        ConnectTime    DateTime `json:"connectTime"`
//        DisconnectTime DateTime `json:"disconnectTime"`
//    }{
//        TmpJSON:        (TmpJSON)(cdr),
//        ConnectTime:    DateTime(cdr.ConnectTime),
//        DisconnectTime: DateTime(cdr.DisconnectTime),
//    })
//}

//func (t DateTime) MarshalJSON() ([]byte, error) {
//	return ([]byte)(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
//}
//
