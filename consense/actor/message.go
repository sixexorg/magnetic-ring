package actor

import "github.com/sixexorg/magnetic-ring/core/mainchain/types"

type StartConsensus struct{}
type StopConsensus struct{}

//internal Message
type TimeOut struct{}
type BlockCompleted struct {
	Block *types.Block
}
