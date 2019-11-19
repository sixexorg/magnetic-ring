package validation

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

var (
	AccountNonceInstance *AccountNonceCache
	autoNonce            map[types.TransactionType]struct{}
)

func init() {
	autoNonce = make(map[types.TransactionType]struct{})
	autoNonce[types.RaiseUT] = struct{}{}
}

func AutoNonceContains(typ types.TransactionType) bool {
	_, ok := autoNonce[typ]
	return ok
}

type AccountNonceCache struct {
	accountNonce map[common.Address]uint64
	height       uint64
	nonceFunc    GetNonceFunc
}
type GetNonceFunc func(height uint64, account common.Address) uint64

func (this *AccountNonceCache) GetAccountNonce(account common.Address) uint64 {
	if this.accountNonce[account] != 0 {
		return this.accountNonce[account]
	}
	return this.nonceFunc(this.height, account)
}

func (this *AccountNonceCache) reset(height uint64) {
	this.accountNonce = make(map[common.Address]uint64)
	this.height = height
}
func NewAccountNonceCache(f GetNonceFunc) {
	AccountNonceInstance = &AccountNonceCache{
		nonceFunc: f,
	}
}
func (this *AccountNonceCache) SetNonce(account common.Address, nonce uint64) {
	this.accountNonce[account] = nonce
}

func (this *AccountNonceCache) nonceSub(account common.Address) {
	this.accountNonce[account]--
}
