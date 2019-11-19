package storelaw

import (
	"io"

	"math/big"

	"sort"

	"github.com/sixexorg/magnetic-ring/common"
)

//AccountStater
type AccountStater interface {
	Hash() common.Hash
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
	Serialization(w io.Writer) error
	GetKey() []byte
	Account() common.Address
	League() common.Address
	BalanceAdd(num *big.Int)
	BalanceSub(num *big.Int)
	EnergyAdd(num *big.Int)
	EnergySub(num *big.Int)
	NonceSet(num uint64)
	Balance() *big.Int
	Energy() *big.Int
	Nonce() uint64
	BonusHeight() uint64
	BonusHSet(num uint64)
	HeightSet(height uint64)
	common.FTreer
}

type AccountStaters []AccountStater

func (s AccountStaters) Len() int           { return len(s) }
func (s AccountStaters) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s AccountStaters) Less(i, j int) bool { return s[i].Hash().String() < s[j].Hash().String() }
func (s AccountStaters) GetHashRoot() common.Hash {
	sort.Sort(s)
	hashes := make([]common.Hash, 0, s.Len())
	for _, v := range s {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}
func (s AccountStaters) GetAccounts() []common.Address {
	addrs := make([]common.Address, 0, s.Len())
	for _, v := range s {
		addrs = append(addrs, v.Account())
	}
	return addrs
}
