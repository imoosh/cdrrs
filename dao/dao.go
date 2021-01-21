package dao

import (
    "centnet-cdrrs/conf"
    "centnet-cdrrs/common/database/orm"
    "gorm.io/gorm"
)

type Dao struct {
    c  *conf.Config
    db *gorm.DB
}

func New(c *conf.Config) (dao *Dao) {
    dao = &Dao{c: c}
    if c.ORM != nil {
        dao.db = orm.NewMySQL(c.ORM)
    }

    return
}
