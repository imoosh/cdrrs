package dao

import (
    "centnet-cdrrs/common/cache/redis"
    "centnet-cdrrs/common/database/orm"
    "centnet-cdrrs/conf"
    "gorm.io/gorm"
)

type Dao struct {
    c     *conf.Config
    db    *gorm.DB
    redis *redis.Pool
}

func New(c *conf.Config) (dao *Dao) {
    dao = &Dao{c: c}
    if c.ORM != nil {
        dao.db = orm.NewMySQL(c.ORM)
    }

    dao.redis = redis.NewPool(c.Redis)

    return
}
