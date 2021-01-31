package model

import (
	"errors"
	"strings"
)

var (
	errUnresolvable         = errors.New("unresolvable number")
	errNotFound             = errors.New("phone position not found")
	fixedNumberPositionMap  map[string]PhonePosition
	mobileNumberPositionMap map[string]PhonePosition
)

// json序列化只需要省市字段
type PhonePosition struct {
	//Id         int    `json:"-"`
	//Prefix     string `json:"-"`
	Phone    string `json:"-"`
	Province string `json:"-"`
	//ProvinceId int64  `json:"-"`
	City string `json:"-"`
	//CityId     int64  `json:"-"`
	//Isp        string `json:"-"`
	Code1 string `json:"-"`
	//Zip        string `json:"-"`
	//Types      string `json:"-"`
}

func (p *PhonePosition) Parse(num string) (err error) {
	length := len(num)
	if !validateNumberString(num) || length < 11 {
		err = errUnresolvable
		return
	}

	if strings.HasPrefix(num, "1") {
		*p, err = QueryMobileNumberPosition(num)
	} else if strings.HasPrefix(num, "0") {
		*p, err = QueryFixedNumberPosition(num)
	} else if strings.HasPrefix(num, "852") {
		*p, err = PhonePosition{Province: "香港", City: "香港"}, nil
	} else if strings.HasPrefix(num, "853") {
		*p, err = PhonePosition{Province: "澳门", City: "澳门"}, nil
	} else {
		err = errUnresolvable
	}

	return err
}

func validateNumberString(num string) bool {
	for _, v := range num {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func (PhonePosition) TableName() string {
	return "phone_position"
}

func CacheNumberPositions(mp, fp map[string]PhonePosition) {
	fixedNumberPositionMap = fp
	mobileNumberPositionMap = mp
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
			return pp, nil
		}
	}

	// 若长度为12，尝试获取前四位归属
	if pp, ok := fixedNumberPositionMap[num[0:4]]; ok {
		return pp, nil
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
		return pp, nil
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
