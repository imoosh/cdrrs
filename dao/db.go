package dao

import (
	"centnet-cdrrs/library/log"
	"errors"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

var (
	asyncDao                        AsyncDao
	emptyCDR                        = VoipRestoredCdr{}
	fixedPhoneNumberAttributionMap  = make(map[string]interface{})
	mobilePhoneNumberAttributionMap = make(map[string]interface{})
	cdrPool                         = sync.Pool{New: func() interface{} { return &VoipRestoredCdr{} }}
)

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

func (ad *AsyncDao) needToCreateTable(cdr, last time.Time) bool {
	// 分表名格式为voip_restored_cdr_YYYYMMDD，cdrCreateTime格式：2006-01-02 15:04:05，lastTableSuffix格式为：2006-01-02
	// 判断当前cdr创建日期是否与上次操作分表时间相同，不同则新建表
	return !(cdr.Year() == last.Year() && cdr.Month() == last.Month() && cdr.Day() == last.Day())
	//return !(cdr.Unix() == last.Unix())
}

func (ad *AsyncDao) Run() {
	var cache []*VoipRestoredCdr
	duration := time.Duration(ad.c.FlushPeriod) * time.Second
	ticker := time.NewTicker(duration)

	lastTime := time.Time{}
	curTableName := "voip_restored_cdr_" + time.Now().Format("20060102")

	go func() {
		for {
			select {
			case m := <-ad.msgQ:
				// 如需要创建新表，先建表，再刷新缓存到旧表
				if ad.needToCreateTable(m.CreateTimeX, lastTime) {
					lastTime = time.Now()
					curTableName = "voip_restored_cdr_" + lastTime.Format("20060102")
					//CreateTable("voip_restored_cdr_" + m.CreateTimeX.Format("20060102"))
					CreateTable(curTableName)
					// 将前一天的部分话单先入旧库
					MultiInsertCDR(curTableName, cache)
					cache = cache[:0]
				}

				cache = append(cache, m)
				if len(cache) == ad.c.MaxFlushCapacity {
					MultiInsertCDR(curTableName, cache)
					cache = cache[:0]
				}

			case <-ticker.C:
				MultiInsertCDR(curTableName, cache)
				cache = cache[:0]
			}
		}
	}()
}

func (ad *AsyncDao) LogCDR(cdr *VoipRestoredCdr) {
	ad.msgQ <- cdr
}

type VoipRestoredCdr struct {
	Id int64 `json:"id"`
	//CallId         string    `json:"callId"`
	CallerIp       string    `json:"callerIp"`
	CallerPort     int       `json:"callerPort"`
	CalleeIp       string    `json:"calleeIp"`
	CalleePort     int       `json:"calleePort"`
	CallerNum      string    `json:"callerNum"`
	CalleeNum      string    `json:"calleeNum"`
	CallerDevice   string    `json:"callerDevice"`
	CalleeDevice   string    `json:"calleeDevice"`
	CalleeProvince string    `json:"calleeProvince"`
	CalleeCity     string    `json:"calleeCity"`
	ConnectTime    int64     `json:"connectTime"`
	DisconnectTime int64     `json:"disconnectTime"`
	Duration       int       `json:"duration"`
	FraudType      string    `json:"fraudType"`
	CreateTime     string    `json:"createTime"`
	CreateTimeX    time.Time `json:"-" orm:"-"`
}

func NewCDR() *VoipRestoredCdr {
	return cdrPool.Get().(*VoipRestoredCdr)
}

func (cdr *VoipRestoredCdr) Free() {
	*cdr = emptyCDR
	cdrPool.Put(cdr)
}

// json序列化只需要省市字段
type PhonePosition struct {
	Id         int    `json:"-"`
	Prefix     string `json:"-"`
	Phone      string `json:"-"`
	Province   string `json:"p"`
	ProvinceId int64  `json:"-"`
	City       string `json:"c"`
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
