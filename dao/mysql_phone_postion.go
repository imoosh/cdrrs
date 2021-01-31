package dao

import (
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/model"
)

func (d *Dao) CacheNumberPositions() (mp, fp map[string]model.PhonePosition) {
	var err error

	fp, err = d.CacheFixedPosition()
	if err != nil {
		log.Error(err)
		return
	}

	mp, err = d.CacheMobilePosition()
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func (d *Dao) CacheFixedPosition() (map[string]model.PhonePosition, error) {
	var pps []model.PhonePosition

	// mysql 5.7查询使用group by报错
	d.db.Exec("SET SESSION sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY,',''));")

	// 查询并缓存座机号码归属地列表
	db := d.db.Model(&model.PhonePosition{})
	db.Select("Code1", "Province", "City").Group("code1").Find(&pps)

	mp := make(map[string]model.PhonePosition, len(pps))
	for _, val := range pps {
		mp[val.Code1] = val
	}
	log.Debugf("fixed  positions cached %d items", len(mp))

	return mp, nil
}

func (d *Dao) CacheMobilePosition() (map[string]model.PhonePosition, error) {
	var pps []model.PhonePosition

	// 查询并缓存手机号码归属地列表
	db := d.db.Model(&model.PhonePosition{})
	db.Select("Phone", "Province", "City").Where("phone IS NOT NULL").Find(&pps)

	mp := make(map[string]model.PhonePosition, len(pps))
	for _, val := range pps {
		mp[val.Phone] = val
	}
	log.Debugf("mobile positions cached %d items", len(mp))
	//fmt.Printf("mp cached %d items\n", len(mp))

	return mp, nil
}
