package dao

import (
    "centnet-cdrrs/common/log"
)

var (
    fixedNumberPositionMap  = make(map[string]interface{})
    mobileNumberPositionMap = make(map[string]interface{})
)

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

func (PhonePosition) TableName() string {
    return "phone_position"
}

func (d *Dao) CachePosition() error {
    var err error

    if err = d.CachePhonePosition(); err != nil {
        return err
    }

    if err = d.CacheFixedPosition(); err != nil {
        return err
    }

    return nil
}

func (d *Dao) CachePhonePosition() error {
    var pps []PhonePosition

    // 查询并缓存座机号码归属地列表
    db := d.db.Model(&PhonePosition{})
    db.Select("Code1", "Province", "City").Group("code1").Find(&pps)
    for _, val := range pps {
        fixedNumberPositionMap[val.Code1] = val
    }
    log.Debugf("fixedNumberPositionMap cached %d items", len(fixedNumberPositionMap))
    //fmt.Printf("fixedNumberPositionMap cached %d items\n", len(fixedNumberPositionMap))

    return nil
}

func (d *Dao) CacheFixedPosition() error {
    var pps []PhonePosition

    // 查询并缓存手机号码归属地列表
    db := d.db.Model(&PhonePosition{})
    db.Select("Phone", "Province", "City").Where("phone IS NOT NULL").Find(&pps)
    for _, val := range pps {
        mobileNumberPositionMap[val.Phone] = val
    }
    log.Debugf("mobileNumberPositionMap cached %d items", len(mobileNumberPositionMap))
    //fmt.Printf("mobileNumberPositionMap cached %d items\n", len(mobileNumberPositionMap))

    return nil
}

// 固话号码获取归属地
func QueryFixedNumberPosition(num string) (PhonePosition, error) {
    length := len(num)
    if length != 11 && length != 12 {
        return PhonePosition{}, errNotFound
    }

    // 若长度为11，先尝试识别前三位归属，再尝试前四位归属
    if length == 11 {
        if pp, ok := fixedNumberPositionMap[num[0:3]]; ok {
            return pp.(PhonePosition), nil
        }
    }

    // 若长度为12，尝试获取前四位归属
    if pp, ok := fixedNumberPositionMap[num[0:4]]; ok {
        return pp.(PhonePosition), nil
    }

    return PhonePosition{}, errNotFound
}

func ValidateFixedNumber(num string) bool {
    length := len(num)
    if length != 11 && length != 12 {
        return false
    }

    // 若长度为11，先尝试识别前三位归属，再尝试前四位归属
    if length == 11 {
        if _, ok := fixedNumberPositionMap[num[0:3]]; ok {
            return true
        }
    }

    // 若长度为12，尝试获取前四位归属
    if _, ok := fixedNumberPositionMap[num[0:4]]; ok {
        return true
    }

    return false
}

// 手机号码获取归属地
func QueryMobileNumberPosition(num string) (PhonePosition, error) {
    if len(num) != 11 {
        //log.Error("invalid num number:", num)
        return PhonePosition{}, errNotFound
    }

    if pp, ok := mobileNumberPositionMap[num[0:7]]; ok {
        return pp.(PhonePosition), nil
    }
    return PhonePosition{}, errNotFound
}

func ValidateMobileNumber(num string) bool {
    if len(num) != 11 {
        return false
    }

    if _, ok := mobileNumberPositionMap[num[0:7]]; ok {
        return true
    }
    return false
}
