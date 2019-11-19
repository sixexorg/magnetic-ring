package votepool

import (
	"sync"

	"github.com/sixexorg/magnetic-ring/common"
)

var ExtVotePool = &extVotePool{
	pool:    make(map[uint64]map[common.Hash]struct{}, 1024),
	prepare: make(map[uint64][]common.Hash, 3),
}

//it only pushes in votes that require a response to a failed result
type extVotePool struct {
	//Key:endHeight val:txHash
	pool    map[uint64]map[common.Hash]struct{}
	prepare map[uint64][]common.Hash
	m       sync.Mutex
}
