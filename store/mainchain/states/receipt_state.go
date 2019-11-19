package states

import (
	"github.com/sixexorg/magnetic-ring/common"
)

type Receipt struct {
	TxHash   common.Hash
	EnergyUsed uint64
}
