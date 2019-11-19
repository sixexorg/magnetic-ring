package account_level

import (
	"encoding/gob"
)

func init() {
	gob.Register(HeightLevel{})
}

type HeightLevel struct {
	Height uint64
	Lv     EasyLevel
}

func (this *HeightLevel) GetHeight() uint64 {
	return this.Height
}
func (this *HeightLevel) GetVal() interface{} {
	return this.Lv
}
