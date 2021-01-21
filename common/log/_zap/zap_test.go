package _zap

import "testing"

func TestLogger(t *testing.T) {
	for i := 0; i < 100000; i++ {
		Infof("%06d --simple _zap logger example", i)
	}
}
