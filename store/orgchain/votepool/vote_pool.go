package votepool

import (
	"sync"

	"github.com/sixexorg/magnetic-ring/common"
)

var VotePool = &votePool{
	pool:    make(map[uint64]map[common.Hash]struct{}, 1024),
	prepare: make(map[uint64][]common.Hash, 3),
}

//it only pushes in votes that require a response to a failed result
type votePool struct {
	//Key:endHeight val:txHash
	pool    map[uint64]map[common.Hash]struct{}
	prepare map[uint64][]common.Hash
	m       sync.Mutex
}

func (this *votePool) naturalDeath(height uint64) []common.Hash {
	this.m.Lock()
	defer this.m.Unlock()
	res := this.pool[height]
	hs := make([]common.Hash, 0, len(res))
	for k, _ := range res {
		hs = append(hs, k)
	}
	delete(this.pool, height)
	return hs
}

func (this *votePool) pushNewVote(height uint64, blockHash common.Hash, txHashes []common.Hash) {
	this.m.Lock()
	defer this.m.Unlock()
	hs := make([]common.Hash, 0, len(txHashes)+1)
	hs = append(hs, blockHash)
	hs = append(hs, txHashes...)
	this.prepare[height] = hs
}
func (this *votePool) commit(height uint64, blockHash common.Hash) {
	this.m.Lock()
	defer this.m.Unlock()

}

func (this *votePool) removeHash(hash common.Hash, height uint64) {
	this.m.Lock()
	defer this.m.Unlock()
	delete(this.pool[height], hash)
}
