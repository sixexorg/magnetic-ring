package storages

import (
	"github.com/hashicorp/golang-lru"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

const (
	ACCOUNT_CACHE_SIZE = 100000
)

type AccountCache struct {
	accountCache *lru.ARCCache
}

func NewAccountCache() (*AccountCache, error) {
	accountCache, err := lru.NewARC(ACCOUNT_CACHE_SIZE)
	if err != nil {
		return nil, err
	}
	return &AccountCache{
		accountCache: accountCache,
	}, nil
}

func (this *AccountCache) GetState(address common.Address) *states.AccountState {
	state, ok := this.accountCache.Get(address.ToString())
	if !ok {
		return nil
	}
	return state.(*states.AccountState)
}

func (this *AccountCache) AddState(state *states.AccountState) {
	this.accountCache.Add(state.Address.ToString(), state)
}

func (this *AccountCache) DeleteState(address common.Address) {
	this.accountCache.Remove(address.ToString())
}
