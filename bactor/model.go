package bactor

import (
	"github.com/sixexorg/magnetic-ring/common"
)

type HeightChange struct {
	Height uint64
}
type LeagueRadarCache struct {
	LeagueId  common.Address
	BlockHash common.Hash
	Height    uint64
}

//P2p and consensus switch
type P2pSyncMutex struct {
	Lock bool
}

type ParticiConfig struct {
	BlkNum      uint64
	View        uint32
	ProcNodes   []string
	ObserNodes  []string
	FailsNodes  []string
	PartiRaw    []string
	StarsSorted []string
}

type Teller interface {
	Tell(message interface{})
}

type InitTellerfunc func() (Teller, error)
