package cdr

import (
    "centnet-cdrrs/dao"
    "sync"
)

var (
    emptyCDR = dao.VoipCDR{}
)

type pool struct {
    p *sync.Pool
}

func newPool() *pool {
    return &pool{p: &sync.Pool{New: func() interface{} { return &dao.VoipCDR{} }}}
}

func (p *pool) New() *dao.VoipCDR {
    return p.p.Get().(*dao.VoipCDR)
}

func (p *pool) Free(cdr *dao.VoipCDR) {
    *cdr = emptyCDR
    p.p.Put(cdr)
}
